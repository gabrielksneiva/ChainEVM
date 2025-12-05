.PHONY: help build test clean deploy deps coverage lint fmt vet install-tools terraform-init terraform-plan terraform-apply docker integration-test ci

help:
	@echo "ChainEVM - AWS Lambda for EVM Execution"
	@echo ""
	@echo "Available commands:"
	@echo "  make build            - Build the Lambda function for AWS"
	@echo "  make build-local      - Build for local testing"
	@echo "  make test             - Run all tests"
	@echo "  make test-short       - Run tests in short mode"
	@echo "  make coverage         - Generate test coverage report"
	@echo "  make lint             - Run linters"
	@echo "  make fmt              - Format code"
	@echo "  make vet              - Run go vet"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make deps             - Download and tidy dependencies"
	@echo "  make deploy           - Deploy Lambda to AWS"
	@echo "  make pre-commit       - Run quick checks before commit"
	@echo "  make ci               - Run full CI pipeline"
	@echo "  make docker           - Build Docker image"
	@echo "  make integration-test - Run integration tests"
	@echo "  make terraform-init   - Initialize Terraform"
	@echo "  make terraform-plan   - Plan Terraform changes"
	@echo "  make terraform-apply  - Apply Terraform changes"
	@echo "  make all              - Run fmt, vet, test, and build"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	go mod verify

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w .

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run linters
lint:
	@echo "Running linters..."
	golangci-lint run ./...

# Build Lambda function for AWS (Linux AMD64)
build: deps
	@echo "Building Lambda function for AWS Linux AMD64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bootstrap cmd/lambda/main.go
	@echo "Creating deployment package..."
	zip -r lambda-deployment.zip bootstrap
	@echo "✓ Build complete: lambda-deployment.zip"

# Build for local testing
build-local: deps
	@echo "Building for local environment..."
	go build -o bin/chainevm cmd/lambda/main.go
	@echo "✓ Local build complete: bin/chainevm"

# Run tests
test:
	@echo "Running tests..."
	go test -v -cover ./...

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | tail -1
	@echo "✓ Coverage report generated: coverage.html"

# Run tests with short mode
test-short:
	@echo "Running tests (short mode)..."
	go test -short -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f bootstrap lambda-deployment.zip
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f *.log
	@echo "✓ Clean complete"

# Deploy Lambda function to AWS
deploy: build
	@echo "Deploying Lambda function to AWS..."
	@if [ -z "$(AWS_REGION)" ]; then \
		echo "Using default region: us-east-1"; \
		aws lambda update-function-code \
			--function-name ChainEVM \
			--zip-file fileb://lambda-deployment.zip \
			--region us-east-1; \
	else \
		aws lambda update-function-code \
			--function-name ChainEVM \
			--zip-file fileb://lambda-deployment.zip \
			--region $(AWS_REGION); \
	fi
	@echo "✓ Deployment complete"

# Terraform commands
terraform-init:
	@echo "Initializing Terraform..."
	cd terraform && terraform init

terraform-plan:
	@echo "Planning Terraform changes..."
	cd terraform && terraform plan

terraform-apply:
	@echo "Applying Terraform changes..."
	cd terraform && terraform apply

terraform-destroy:
	@echo "Destroying Terraform resources..."
	cd terraform && terraform destroy

# Run all checks and build
all: fmt vet test build
	@echo "✓ All tasks completed successfully"

# Run quick checks before commit
pre-commit: fmt vet test-short
	@echo "✓ Pre-commit checks passed"

# Run full CI pipeline
ci: deps fmt vet lint test coverage build
	@echo "✓ CI pipeline completed successfully"

# Build Docker image (for local development with LocalStack)
docker:
	@echo "Building Docker image..."
	docker build -t chainevm:latest .
	@echo "✓ Docker image built"

# Run integration tests (requires LocalStack/DynamoDB local)
integration-test:
	@echo "Running integration tests..."
	@echo "Make sure LocalStack or DynamoDB local is running"
	go test -v -race -tags=integration ./...

.DEFAULT_GOAL := help
