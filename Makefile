.PHONY: help build run test clean docker-build docker-run docker-stop docker-logs deps lint fmt

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with docker-compose"
	@echo "  docker-stop  - Stop docker-compose services"
	@echo "  docker-logs  - Show docker-compose logs"

# Build the application
build:
	go build -o smart-mail-relay .

# Run the application locally
run:
	go run .

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f smart-mail-relay
	go clean

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Build Docker image
docker-build:
	docker build -t smart-mail-relay .

# Run with docker-compose
docker-run:
	docker-compose up -d

# Stop docker-compose services
docker-stop:
	docker-compose down

# Show docker-compose logs
docker-logs:
	docker-compose logs -f

# Get Gmail OAuth2 token
get-token:
	go run tools/get_token.go

# Database operations
db-migrate:
	go run . --migrate

# Health check
health:
	curl -f http://localhost:8080/healthz

# API examples
api-examples:
	@echo "Creating a forwarding rule:"
	@echo "curl -X POST http://localhost:8080/api/v1/rules \\"
	@echo "  -H 'Content-Type: application/json' \\"
	@echo "  -d '{\"keyword\": \"urgent\", \"target_email\": \"admin@company.com\", \"enabled\": true}'"
	@echo ""
	@echo "Listing rules:"
	@echo "curl http://localhost:8080/api/v1/rules"
	@echo ""
	@echo "Getting metrics:"
	@echo "curl http://localhost:8080/metrics"

# Development setup
dev-setup: deps fmt lint test
	@echo "Development setup completed"

# Production build
prod-build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o smart-mail-relay .

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest 