BINARY ?= sim-cli

.PHONY: all build run clean fmt vet lint test cover docker-build docker-run help

## Default

all: fmt lint test build

## Build

build:
	@echo "Building binary..."
	CGO_ENABLED=0 go build -o $(BINARY) .

run:
	@echo "Running..."
	go run .

clean:
	@echo "Cleaning..."
	rm -f $(BINARY) coverage.out

## Quality

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...

lint:
	@echo "Running linter..."
	golangci-lint run

## Test

test:
	@echo "Running unit tests..."
	go test -coverprofile=coverage.out ./...

cover:
	@echo "Running tests with coverage..."
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## Docker

docker-build:
	@echo "Building Docker image..."
	docker build --target prod -t $(BINARY) .

docker-run:
	@echo "Running Docker image..."
	docker run --rm $(BINARY)

## Help

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "  all          fmt → lint → test → build"
	@echo "  build        compile the binary"
	@echo "  run          go run ."
	@echo "  clean        remove binary and coverage report"
	@echo "  fmt          go fmt ./..."
	@echo "  vet          go vet ./..."
	@echo "  lint         golangci-lint run"
	@echo "  test         run unit tests with coverage"
	@echo "  cover        run tests and open HTML coverage report"
	@echo "  docker-build build production Docker image"
	@echo "  docker-run   run the Docker image"
