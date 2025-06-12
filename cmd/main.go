package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	healthcheck "github.com/akyriako/typesense-healthcheck"
	"github.com/caarlos0/env/v11"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
)

var (
	config   healthcheck.Config
	logger   *slog.Logger
	hcClient *healthcheck.HealthCheckClient
)

const (
	exitCodeConfigurationError int = 78
)

func init() {
	err := env.Parse(&config)
	if err != nil {
		slog.Error(fmt.Sprintf("parsing env variables failed: %s", err.Error()))
		os.Exit(exitCodeConfigurationError)
	}

	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(config.LogLevel),
	}))

	slog.SetDefault(logger)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hcClient = healthcheck.NewHealthCheckClient(config)

	http.HandleFunc("/livez", livezHandler)
	http.HandleFunc("/readyz", readyzHandler)
	http.HandleFunc("/", indexHandler)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		<-sigChan

		logger.Warn("termination signal received, shutting down gracefully...")
		cancel()
	}()

	server := &http.Server{Addr: fmt.Sprintf(":%d", config.HealthCheckPort)}

	go func() {
		logger.Info("starting typesense healthcheck server...", "port", config.HealthCheckPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(fmt.Sprintf("error starting server: %v", err))
			os.Exit(-1)
		}
	}()

	<-ctx.Done()

	logger.Info("shutting down server...")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error(fmt.Sprintf("error during server shutdown: %v", err))
		os.Exit(-1)
	}
	logger.Info("server shut down successfully")
}

func livezHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("."))
}

func readyzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := http.StatusOK
	internalError := false

	health, err := hcClient.GetClusterHealth(r.Context())
	if err != nil {
		logger.Error(fmt.Sprintf("error getting cluster health: %v", err))
		internalError = true
	}

	body, err := json.Marshal(health)
	if err != nil {
		logger.Error(fmt.Sprintf("error marshalling cluster health: %v", err))
		internalError = true
	}

	if !health.ClusterHealth {
		if internalError {
			status = http.StatusInternalServerError
		} else {
			status = http.StatusServiceUnavailable
		}
	}

	w.WriteHeader(status)
	w.Write(body)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	uiPath := "./ui"
	tmpl, err := template.New("vue.html").ParseFiles(path.Join(uiPath, "vue.html"))
	if err != nil {
		logger.Error(fmt.Sprintf("error parsing template: %v", err))

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(path.Join(uiPath, "vue.html") + err.Error()))
		return
	}

	data := struct {
		Title string
		Logo  string
	}{
		Title: "Typesense Healthcheck",
		Logo:  config.Logo,
	}

	if err := tmpl.Execute(w, data); err != nil {
		logger.Error(fmt.Sprintf("error executing template: %v", err))
		http.Error(w, path.Join(uiPath, "vue.html")+err.Error(), http.StatusInternalServerError)
	}
}
