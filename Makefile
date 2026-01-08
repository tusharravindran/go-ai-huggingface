.PHONY: help build test test-coverage test-race clean lint fmt vet deps dev docker-build docker-run install-tools pre-commit

# Default target
help: ## Show this help message
	@echo 'Usage: make [TARGET]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build variables
BINARY_NAME=server
BINARY_PATH=./cmd/server
BUILD_DIR=./bin
DOCKER_IMAGE=go-ai-huggingface
DOCKER_TAG=latest

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-w -s -X main.version=$(shell git describe --tags --always --dirty)"
BUILD_FLAGS=-v $(LDFLAGS)

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)

build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(BINARY_PATH)

build-darwin: ## Build for macOS
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(BINARY_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(BINARY_PATH)

build-windows: ## Build for Windows
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(BINARY_PATH)

build-all: build-linux build-darwin build-windows ## Build for all platforms

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	$(GOTEST) -v -race ./...

test-benchmark: ## Run benchmark tests
	@echo "Running benchmark tests..."
	$(GOTEST) -bench=. -benchmem ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@$(GOCMD) clean -cache

lint: ## Run golangci-lint
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run 'make install-tools'" && exit 1)
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	$(GOGET) -u ./...

dev: ## Run in development mode with hot reload
	@echo "Starting development server..."
	@which air > /dev/null || (echo "Air not installed. Run 'go install github.com/cosmtrek/air@latest'" && exit 1)
	air

run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME)

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build ## Build and run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 \
		-e HUGGINGFACE_API_KEY=${HUGGINGFACE_API_KEY} \
		-e LOG_LEVEL=debug \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push: docker-build ## Push Docker image to registry
	@echo "Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

install-tools: ## Install development tools
	@echo "Installing development tools..."
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) -u github.com/cosmtrek/air@latest
	$(GOGET) -u github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	$(GOGET) -u golang.org/x/tools/cmd/goimports@latest

security: ## Run security checks
	@echo "Running security checks..."
	@which gosec > /dev/null || (echo "gosec not installed. Run 'make install-tools'" && exit 1)
	gosec ./...

pre-commit: fmt vet lint test-race test-coverage security ## Run all pre-commit checks
	@echo "All pre-commit checks passed!"

ci: deps pre-commit build ## Run CI pipeline locally

# Environment setup
env-example: ## Create .env.example file
	@echo "Creating .env.example..."
	@cat > .env.example << 'EOF'
	# Server Configuration
	SERVER_PORT=8080
	SERVER_HOST=localhost
	SERVER_READ_TIMEOUT=30s
	SERVER_WRITE_TIMEOUT=30s

	# Hugging Face Configuration
	HUGGINGFACE_API_KEY=your-api-key-here
	HUGGINGFACE_BASE_URL=https://api-inference.huggingface.co
	HUGGINGFACE_DEFAULT_MODEL=gpt2
	HUGGINGFACE_TIMEOUT=30s
	HUGGINGFACE_MAX_TOKENS=100
	HUGGINGFACE_TEMPERATURE=0.7

	# Logging Configuration
	LOG_LEVEL=info
	LOG_FORMAT=json
	LOG_STRUCTURED=true

	# Database Configuration (optional)
	# DATABASE_DRIVER=postgres
	# DATABASE_HOST=localhost
	# DATABASE_PORT=5432
	# DATABASE_NAME=ai_service
	# DATABASE_USERNAME=postgres
	# DATABASE_PASSWORD=password
	EOF

# Help for setup
setup: ## Initial project setup
	@echo "Setting up project..."
	$(MAKE) deps
	$(MAKE) install-tools
	$(MAKE) env-example
	@echo ""
	@echo "Setup complete! Next steps:"
	@echo "1. Copy .env.example to .env and fill in your values"
	@echo "2. Get your Hugging Face API key from https://huggingface.co/settings/tokens"
	@echo "3. Run 'make dev' to start development server"
	@echo "4. Visit http://localhost:8080/health to verify the service is running"

# Deployment helpers
deploy-staging: ## Deploy to staging (customize as needed)
	@echo "Deploying to staging..."
	# Add your staging deployment commands here

deploy-prod: ## Deploy to production (customize as needed)
	@echo "Deploying to production..."
	# Add your production deployment commands here

# Database operations (if using database)
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	# Add your migration commands here

db-seed: ## Seed database with test data
	@echo "Seeding database..."
	# Add your seeding commands here

# Monitoring and debugging
logs: ## Show application logs (for containerized deployments)
	@echo "Showing logs..."
	docker logs -f $(shell docker ps -q --filter ancestor=$(DOCKER_IMAGE):$(DOCKER_TAG))

ps: ## Show running processes
	@echo "Running processes:"
	@ps aux | grep $(BINARY_NAME) | grep -v grep || echo "No $(BINARY_NAME) processes found"

# Utility targets
version: ## Show version information
	@echo "Go version: $(shell go version)"
	@echo "Git commit: $(shell git rev-parse --short HEAD)"
	@echo "Git tag: $(shell git describe --tags --always)"
	@echo "Build time: $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")"

info: version ## Show project information
	@echo ""
	@echo "Project: Go AI Hugging Face Service"
	@echo "Binary: $(BINARY_NAME)"
	@echo "Build dir: $(BUILD_DIR)"
	@echo "Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"