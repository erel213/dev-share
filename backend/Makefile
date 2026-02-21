.PHONY: help test test-integration test-workspace test-user test-unit test-all clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test-integration: ## Run all integration tests
	@./scripts/run-integration-tests.sh all

test-integration-v: ## Run all integration tests with verbose output
	@./scripts/run-integration-tests.sh -v all

test-workspace: ## Run workspace integration tests only
	@./scripts/run-integration-tests.sh workspace

test-workspace-v: ## Run workspace integration tests with verbose output
	@./scripts/run-integration-tests.sh -v workspace

test-user: ## Run user integration tests only
	@./scripts/run-integration-tests.sh user

test-user-v: ## Run user integration tests with verbose output
	@./scripts/run-integration-tests.sh -v user

test-unit: ## Run unit tests
	@go test ./internal/... ./pkg/... -v

test-all: test-unit test-integration ## Run all tests (unit + integration)

test-keep: ## Run integration tests and keep containers running
	@./scripts/run-integration-tests.sh -k all

docker-up: ## Start test environment only (no tests)
	@docker compose -f docker-compose.test.yml up -d --build

docker-down: ## Stop test environment
	@docker compose -f docker-compose.test.yml down -v

docker-logs: ## Show backend test logs
	@docker compose -f docker-compose.test.yml logs -f backend-test

clean: docker-down ## Clean up test environment and temporary files
	@echo "Cleaning up..."

build: ## Build the backend
	@go build -o bin/server ./cmd/server

run: ## Run the backend locally
	@go run ./cmd/server/main.go

lint: ## Run linters
	@go vet ./...
	@go fmt ./...

deps: ## Download dependencies
	@go mod download
	@go mod tidy

.DEFAULT_GOAL := help
