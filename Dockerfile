# Stage 1: Build
FROM golang:1.23 AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o typesense-healthcheck ./cmd

# Stage 2: Package
FROM alpine:3.18

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/typesense-healthcheck /app/typesense-healthcheck

# Copy the UI directory so ui/vue.html is available at /app/ui/vue.html
COPY --from=builder /app/ui ./ui

EXPOSE 8808

CMD ["/app/typesense-healthcheck"]
