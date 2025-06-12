# Configuration variables
REGISTRY ?= akyriako78#$(shell docker info | sed '/Username:/!d;s/.* //')
IMAGE_NAME ?= typesense-healthcheck
TAG ?= 0.1.3
DOCKERFILE ?= Dockerfile

# Build binary
build:
	@echo "Building Go binary..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cmd/typesense-healthcheck ./cmd

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(REGISTRY)/$(IMAGE_NAME):$(TAG) -f $(DOCKERFILE) .

# Push Docker image
docker-push:
	@echo "Pushing Docker image to registry..."
	docker push $(REGISTRY)/$(IMAGE_NAME):$(TAG)

# Clean up
clean:
	@echo "Cleaning up..."
	rm -f cmd/typesense-healthcheck

# Default target
.PHONY: build docker-build docker-push clean
