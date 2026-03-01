# Makefile for the Goly application

# --- Variables ---
# These variables can be overridden from the command line.
# Example: make DOCKER_IMAGE=my-registry/my-goly-app docker-build
APP_NAME=goly-app
DOCKER_IMAGE?=$(APP_NAME):latest
DOCKER_REGISTRY?=your-docker-registry # Replace with your Docker registry

# --- Go settings ---
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_CLEAN=$(GO_CMD) clean
GO_TEST=$(GO_CMD) test
GO_FILES=$(wildcard *.go)

# --- Targets ---

.PHONY: all build test docker-build docker-push deploy clean help

all: build

# Build the Go application
build:
	@echo "Building the application..."
	$(GO_BUILD) -o $(APP_NAME) goly/main.go

# Run the unit tests
test:
	@echo "Running unit tests..."
	$(GO_TEST) -v ./...

# Build the Docker image
docker-build:
	@echo "Building Docker image: $(DOCKER_IMAGE)"
	@docker build -t $(DOCKER_IMAGE) .

# Push the Docker image to a registry
# You must be logged in to your Docker registry for this to work.
docker-push:
	@echo "Pushing Docker image to $(DOCKER_REGISTRY)"
	@docker tag $(DOCKER_IMAGE) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)

# Deploy the application to Kubernetes using Terraform
# This is the "single command" to deploy the application.
# It assumes you have configured your Terraform variables in a .tfvars file.
deploy: docker-build docker-push
	@echo "Deploying to Kubernetes..."
	cd infra/terraform && terraform apply -auto-approve

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	$(GO_CLEAN)
	rm -f $(APP_NAME)

# Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all           Build the application (default)"
	@echo "  build         Build the Go application"
	@echo "  test          Run the unit tests"
	@echo "  docker-build  Build the Docker image"
	@echo "  docker-push   Push the Docker image to a registry"
	@echo "  deploy        Deploy the application to Kubernetes"
	@echo "  clean         Clean up build artifacts"
	@echo "  help          Display this help message"
