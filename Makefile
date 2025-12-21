# Simple Makefile for a Go project

# Build the application
all: build

build:
	@echo "Building..."
	@go build -o main ./cmd

# Run the application
run:
	@if lsof -ti :8080 > /dev/null 2>&1; then \
		echo "Port 8080 is already in use. Killing existing process..."; \
		lsof -ti :8080 | xargs kill -9 2>/dev/null || true; \
		sleep 1; \
	fi
	@go run ./cmd &
	@cd frontend && pnpm install --prefer-offline
	@cd frontend && pnpm run dev
# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Database migrations
migrate-up:
	@set -a && [ -f .env ] && source .env; set +a && \
	goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	@set -a && [ -f .env ] && source .env; set +a && \
	goose -dir migrations postgres "$$DATABASE_URL" down

migrate-status:
	@set -a && [ -f .env ] && source .env; set +a && \
	goose -dir migrations postgres "$$DATABASE_URL" status

migrate-create:
	@read -p "Enter migration name: " name; \
	goose -dir migrations create $$name sql

# Generate sqlc code
sqlc-generate:
	@if command -v sqlc > /dev/null; then \
		sqlc generate; \
		echo "sqlc code generated successfully"; \
	else \
		echo "sqlc is not installed. Install it with: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"; \
		exit 1; \
	fi

.PHONY: all build run clean watch docker-run docker-down migrate-up migrate-down migrate-status migrate-create sqlc-generate
