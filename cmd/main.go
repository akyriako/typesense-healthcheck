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

	health, err := hcClient.GetClusterHealth(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error getting cluster health"))
		return
	}

	body, err := json.Marshal(health)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error marshalling cluster health"))
		return
	}

	status := http.StatusOK
	if !health.ClusterHealth {
		status = http.StatusServiceUnavailable
	}

	w.WriteHeader(status)
	w.Write(body)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	uiPath := "./ui"
	tmpl, err := template.New("vue.html").ParseFiles(path.Join(uiPath, "vue.html"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(path.Join(uiPath, "vue.html") + err.Error()))
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	data := struct {
		Title    string
		Logo     string
		ServedBy string
	}{
		Title:    "Typesense Healthcheck",
		Logo:     config.Logo,
		ServedBy: hostname,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, path.Join(uiPath, "vue.html")+err.Error(), http.StatusInternalServerError)
	}
}
