.PHONY: help cover run lint test build docker-build docker-up docker-down docker-logs setup clean

help:
	@echo "ChainConnector - Blockchain RPC Service"
	@echo ""
	@echo "Available targets:"
	@echo "  help              - Show this help message"
	@echo "  test              - Run all tests"
	@echo "  cover             - Run tests with coverage report"
	@echo "  run               - Run service locally (requires PostgreSQL)"
	@echo "  build             - Build the service binary"
	@echo "  lint              - Run linters and formatters"
	@echo "  vet               - Run go vet"
	@echo "  fmt               - Format code"
	@echo "  clean             - Remove build artifacts"
	@echo ""
	@echo "Docker targets:"
	@echo "  setup             - Initialize Docker environment (Anvil, PostgreSQL, ChainConnector)"
	@echo "  docker-build      - Build Docker image"
	@echo "  docker-up         - Start all Docker services"
	@echo "  docker-down       - Stop all Docker services"
	@echo "  docker-logs       - View logs from all services"
	@echo "  docker-sepolia-up - Start local Sepolia node"
	@echo "  docker-sepolia-down - Stop local Sepolia node"
	@echo "  docker-sepolia-logs - View Sepolia node logs"
	@echo "  docker-clean      - Remove containers and volumes"

test:
	go test ./... -v -count=1
.PHONY: test

cover:
	go test ./... -v -count=1 -covermode=atomic -coverprofile=coverage.out && go tool cover -func=coverage.out && rm coverage.out

run:
	go run cmd/chainconnector/main.go

build:
	CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o chainconnector ./cmd/chainconnector

lint:
	golangci-lint run --timeout 5m && go fmt ./... && go vet ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -f chainconnector coverage.out

setup:
	@chmod +x scripts/dev-setup.sh
	@scripts/dev-setup.sh

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-sepolia-up:
	docker-compose -f docker-compose.sepolia.yml up -d

docker-sepolia-down:
	docker-compose -f docker-compose.sepolia.yml down

docker-sepolia-logs:
	docker-compose -f docker-compose.sepolia.yml logs -f
