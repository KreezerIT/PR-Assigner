.PHONY: build run test coverage docker-up docker-down docker-test migrate-create migrate-up migrate-down lint deps clean

include .env
export

.DEFAULT_GOAL := help

help:
	@echo Available commands:
	@echo   docker-up          - Start services
	@echo   docker-down        - Stop services
	@echo   docker-test        - Run all tests in Docker
	@echo   build              - Build application
	@echo   run                - Run application
	@echo   test               - Run unit tests (with coverage)
	@echo   test-integration   - Run integration tests
	@echo   coverage           - Generate coverage report
	@echo   migrate-create     - Create migration (NAME=name)
	@echo   migrate-up         - Apply migrations
	@echo   migrate-down       - Rollback migration
	@echo   lint               - Run linter
	@echo   deps               - Install dependencies
	@echo   clean              - Clean generated files

build: ## Build application
	go build -o bin/api cmd/api/main.go

run: ## Run application locally
	go run cmd/api/main.go

test: ## Run unit-tests
	go test -coverprofile="coverage.out" ./internal/...

test-integration: ## Run integration tests
	@docker-compose -f docker-compose.test.yml up -d postgres_test
	@timeout /t 3 /nobreak
	set "TEST_DB_HOST=localhost" && set "TEST_DB_PORT=5433" && set "TEST_DB_USER=prservice_test" && set "TEST_DB_PASSWORD=test_pass" && set "TEST_DB_NAME=pr_reviewer_test" && go test -v -tags=integration ./tests/integration/...
	@docker-compose -f docker-compose.test.yml down

coverage: test ## Show test coverage
	go tool cover -html=coverage.out -o coverage.html

docker-up: ## Start all services with docker-compose
	docker-compose up --build -d

docker-down: ## Stop all services
	docker-compose down -v

docker-test: ## Run all tests in docker
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down -v

migrate-create: ## Create new migration (usage: make migrate-create NAME=migration_name)
	migrate create -ext sql -dir migrations -seq $(NAME)

migrate-up: ## Apply migrations
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

migrate-down: ## Rollback last migration
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down 1

lint: ## Run golangci-lint
	golangci-lint run --timeout 5m

deps: ## Install dependencies
	go mod download; go mod tidy

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache -testcache
