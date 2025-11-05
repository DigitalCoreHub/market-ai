.PHONY: help dev build test clean docker-up docker-down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Run in development mode
	go run cmd/server/main.go

build: ## Build the application
	go build -o bin/marketai cmd/server/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build files
	rm -rf bin/

docker-up: ## Start Docker services
	docker-compose up -d

docker-down: ## Stop Docker services
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

install: ## Install dependencies
	go mod download
	go mod tidy

lint: ## Run linter
	golangci-lint run

fmt: ## Format code (gofmt, gofumpt, goimports) and tidy modules
	@echo "-> Running go fmt"
	go fmt ./...
	@echo "-> Running gofumpt (if installed)"
	@gofumpt -w . 2>/dev/null || true
	@echo "-> Running goimports (if installed)"
	@goimports -w . 2>/dev/null || true
	@echo "-> Running go mod tidy"
	go mod tidy

fmt-check: ## Fail if files are not gofmt'ed
	@out=$$(gofmt -s -l .); if [ -n "$$out" ]; then echo "Unformatted files:"; echo "$$out"; exit 1; fi
