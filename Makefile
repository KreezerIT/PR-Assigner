.PHONY: build run test test-unit test-integration docker-up docker-down migrate-up migrate-down lint coverage docker-test

build: ## Build application
	go build -o bin/api cmd/api/main.go

run: ## Run application locally
	go run cmd/api/main.go

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	go test -coverprofile="coverage.out" ./internal/...

test-integration: ## Run integration tests
	@echo "Starting test database"
	@docker-compose -f docker-compose.test.yml up -d postgres_test
	@timeout /t 3 /nobreak
	@echo "Running integration tests"
	set "TEST_DB_HOST=localhost" && set "TEST_DB_PORT=5433" && set "TEST_DB_USER=prservice_test" && set "TEST_DB_PASSWORD=test_pass" && set "TEST_DB_NAME=pr_reviewer_test" && go test -v -tags=integration ./tests/integration/...
	@docker-compose -f docker-compose.test.yml down

coverage: test-unit ## Show test coverage
	go tool cover -html=coverage.out

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
	migrate -path migrations -database "postgres://prservice:prservice_pass@localhost:5432/pr_reviewer?sslmode=disable" up

migrate-down: ## Rollback last migration
	migrate -path migrations -database "postgres://prservice:prservice_pass@localhost:5432/pr_reviewer?sslmode=disable" down 1

lint: ## Run golangci-lint
	golangci-lint run --timeout 5m

deps: ## Install dependencies
	go mod download; go mod tidy

.DEFAULT_GOAL := help