.PHONY: build run test clean docker-build docker-run help

# Variables
APP_NAME=goarchive
DOCKER_IMAGE=goarchive:latest
GO_FILES=$(shell find . -name '*.go' -type f)

# Default target
.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the application binary
build:
	@echo "Building $(APP_NAME)..."
	cd cmd/goarchive && go build -o ../../$(APP_NAME) .

## run: Run the application locally
run:
	@echo "Running $(APP_NAME)..."
	cd cmd/goarchive && go run main.go

## test: Run tests
test:
	@echo "Running tests..."
	go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## lint: Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

## fmt: Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofumpt -l -w .

## tidy: Tidy and verify module dependencies
tidy:
	@echo "Tidying dependencies..."
	@echo "- Core module..."
	go mod tidy
	go mod verify
	@echo "- Provider modules..."
	cd database/postgres && go mod tidy
	cd storage/disk && go mod tidy
	cd storage/s3 && go mod tidy
	@echo "- CLI module..."
	cd cmd/goarchive && go mod tidy
	@echo "- Example modules..."
	cd examples/basic-backup && go mod tidy
	cd examples/using-env-config && go mod tidy
	@echo "✓ All modules tidied (verify skipped for modules with local replaces)"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(APP_NAME)
	rm -f coverage.out coverage.html
	go clean

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

## docker-build-scheduler: Build Docker image with scheduler
docker-build-scheduler:
	@echo "Building Docker scheduler image..."
	docker build -f Dockerfile.scheduler -t goarchive:scheduler .

## docker-run: Run the application in Docker
docker-run: docker-build
	@echo "Running $(APP_NAME) in Docker..."
	docker run --rm \
		--env-file .env \
		$(DOCKER_IMAGE)

## docker-compose-up: Start services with docker-compose
docker-compose-up:
	@echo "Starting services..."
	docker-compose up -d

## docker-compose-down: Stop services with docker-compose
docker-compose-down:
	@echo "Stopping services..."
	docker-compose down

## docker-compose-test: Run full test with docker-compose
docker-compose-test:
	@echo "Running full test with docker-compose..."
	docker-compose up -d postgres localstack
	@echo "Waiting for services to be ready..."
	sleep 10
	@echo "Creating S3 bucket..."
	docker-compose run --rm goarchive sh -c '\
		aws --endpoint-url=http://localstack:4566 s3 mb s3://backups --region us-east-1 || true'
	@echo "Running backup..."
	docker-compose run --rm goarchive
	docker-compose down

## install-tools: Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
