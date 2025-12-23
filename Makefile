.PHONY: build test clean lint run help

BINARY_NAME=slurm_exporter
BUILD_DIR=bin
VERSION?=dev
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildTime=${BUILD_TIME}"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ./cmd/slurm_exporter
	@echo "Build complete: ${BUILD_DIR}/${BINARY_NAME}"

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

coverage: test ## Run tests and show coverage
	go tool cover -html=coverage.out

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf ${BUILD_DIR}
	rm -f coverage.out
	rm -rf dist/
	@echo "Clean complete"

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	golangci-lint run ./...

run: build ## Build and run the exporter
	${BUILD_DIR}/${BINARY_NAME} --config.file configs/config.yaml

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

install: build ## Install the binary
	@echo "Installing ${BINARY_NAME}..."
	go install ${LDFLAGS} ./cmd/slurm_exporter

.DEFAULT_GOAL := help
