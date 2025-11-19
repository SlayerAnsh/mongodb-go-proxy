.PHONY: swagger run build clean deps docker-build docker-up docker-down docker-logs docker-restart docker-clean docker-shell

# Install swag if not already installed
swagger-install:
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger documentation
swagger: swagger-install
	@swag init -g main.go -o docs

# Run the server
run: swagger
	@go run main.go

# Build the application
build:
	@go build -o bin/server main.go

# Clean generated files
clean:
	@rm -rf bin/
	@rm -rf docs/

# Install dependencies
deps:
	@go mod download
	@go mod tidy

# Docker commands
docker-build:
	@echo "Building Docker image..."
	@docker compose build

docker-up:
	@echo "Starting Docker containers..."
	@docker compose up -d --build

docker-down:
	@echo "Stopping Docker containers..."
	@docker compose down

docker-logs:
	@docker compose logs -f

docker-restart:
	@echo "Restarting Docker containers..."
	@docker compose restart

docker-clean:
	@echo "Stopping containers and removing volumes..."
	@docker compose down -v
	@echo "Removing Docker images..."
	@docker compose rm -f

docker-shell:
	@docker compose exec server sh

# Build and run with Docker
docker-run: docker-build docker-up
	@echo "Containers are running. Use 'make docker-logs' to view logs."

# Rebuild and restart
docker-rebuild: docker-down docker-build docker-up
	@echo "Containers rebuilt and restarted."

