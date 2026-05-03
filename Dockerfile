# Stage 1: Build
FROM golang:1.24 AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
ARG BUILT_BY=local
ARG DIRTY=unknown

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
  -ldflags="-s -w \
  -X github.com/akyriako/typesense-healthcheck/internal/version.Version=${VERSION} \
  -X github.com/akyriako/typesense-healthcheck/internal/version.Commit=${COMMIT} \
  -X github.com/akyriako/typesense-healthcheck/internal/version.Date=${DATE} \
  -X github.com/akyriako/typesense-healthcheck/internal/version.BuiltBy=${BUILT_BY} \
  -X github.com/akyriako/typesense-healthcheck/internal/version.Dirty=${DIRTY}" \
  -o typesense-healthcheck ./cmd

# Stage 2: Package
FROM alpine:3.21

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/typesense-healthcheck /app/typesense-healthcheck

# Copy the UI directory so ui/vue.html is available at /app/ui/vue.html
COPY --from=builder /app/ui ./ui

EXPOSE 8808

CMD ["/app/typesense-healthcheck"]
