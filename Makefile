# Configuration variables
REGISTRY ?= akyriako78#$(shell docker info | sed '/Username:/!d;s/.* //')
IMAGE_NAME ?= typesense-healthcheck
TAG ?= 0.2.0
DOCKERFILE ?= Dockerfile
PLATFORMS ?= linux/amd64,linux/arm64,linux/s390x,linux/ppc64le
DOCKERX_BUILDER ?= typesense-healthcheck-builder

ITERATION ?= 1
TARGET_ENV ?= dev
VERSION ?= $(TAG)-$(TARGET_ENV).$(ITERATION)
COMMIT  ?= $(shell git rev-parse --short HEAD)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILT_BY ?= local
DIRTY   ?= $(shell test -n "$$(git status --porcelain)" && echo true || echo false)

LDFLAGS := -s -w \
	-X 'github.com/akyriako/typesense-healthcheck/internal/version.Version=$(VERSION)' \
	-X 'github.com/akyriako/typesense-healthcheck/internal/version.Commit=$(COMMIT)' \
	-X 'github.com/akyriako/typesense-healthcheck/internal/version.Date=$(DATE)' \
	-X 'github.com/akyriako/typesense-healthcheck/internal/version.BuiltBy=$(BUILT_BY)' \
	-X 'github.com/akyriako/typesense-healthcheck/internal/version.Dirty=$(DIRTY)'

build:
	@echo "Building Go binary with flags..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o cmd/typesense-healthcheck ./cmd

## Build binary
#build:
#	@echo "Building Go binary..."
#	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cmd/typesense-healthcheck ./cmd

# Build a docker builder
docker-builder:
	@echo "Creating buildx builder..."
	docker buildx create --name ${DOCKERX_BUILDER} || true
	docker buildx inspect --builder ${DOCKERX_BUILDER} --bootstrap

# Build Docker image
docker-build: docker-builder
	@echo "Building Docker image..."
	docker buildx build --load --builder ${DOCKERX_BUILDER} -t $(REGISTRY)/$(IMAGE_NAME):$(TAG) -f $(DOCKERFILE) .

# Push Docker image
docker-push: docker-builder
	@echo "Pushing Docker image to registry..."
	docker buildx build --push --builder ${DOCKERX_BUILDER} --platform ${PLATFORMS}  -t $(REGISTRY)/$(IMAGE_NAME):$(TAG) -f $(DOCKERFILE) .

# Clean up
clean:
	@echo "Cleaning up..."
	rm -f cmd/typesense-healthcheck

# Default target
.PHONY: build docker-build docker-push clean
