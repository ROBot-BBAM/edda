.PHONY: help up down build clean migrate test

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

up: ## Start all services with Docker Compose
	docker-compose up -d

up-build: ## Build and start all services
	docker-compose up --build -d 

down: ## Stop all services
	docker-compose down

build: ## Build all Docker images
	docker-compose build

clean: ## Remove containers and volumes
	docker-compose down -v

migrate: ## Run database migrations (requires backend to be running)
	@echo "Migrations run automatically on backend startup"

test-backend: ## Run backend tests
	cd backend && go test ./...

test-frontend: ## Run frontend tests
	cd frontend && npm test

dev-backend: ## Run backend locally (requires PostgreSQL running)
	cd backend && go run cmd/server/main.go

dev-frontend: ## Run frontend locally
	cd frontend && npm start
