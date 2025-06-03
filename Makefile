BINARY_NAME_SERVER=alcatraz-live
MAIN_PATH_SERVER=./cmd/server/main.go
CONFIG_FILE=config.yaml
BUILD_DIR=build
BINARY_NAME_SENDER=alcatraz-live-sender
MAIN_PATH_SENDER=./cmd/sender/main.go


GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod


.PHONY: build-server
build-server:  ## Build the application
	@echo "Building $(BINARY_NAME_SERVER)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME_SERVER) $(MAIN_PATH_SERVER)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME_SERVER)"

.PHONY: build-sender
build-sender:  ## Build the application
	@echo "Building $(BINARY_NAME_SENDER)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME_SENDER) $(MAIN_PATH_SENDER)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME_SENDER)"

.PHONY: run-server
run-server: build-server  ## Run the application
	@echo "Running $(BINARY_NAME_SERVER)..."
	./$(BUILD_DIR)/$(BINARY_NAME_SERVER) \
		-listen-address=127.0.0.1 \
		-port=9000 \
		-log-level=debug \
		-log-type=console

.PHONY: run-sender
run-sender: build-sender  ## Run the application
	@echo "Running $(BINARY_NAME_SENDER)..."
	./$(BUILD_DIR)/$(BINARY_NAME_SENDER)

.PHONY: lint
lint:  ## Run golangci-lint
	@echo "Running golangci-lint..."
	golangci-lint -c .golangci.yml run ./...

.PHONY: test
test:  ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

.PHONY: test-coverage
test-coverage:  ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: clean
clean:  ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

.PHONY: tidy
tidy:  ## Tidy and verify dependencies
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify

.PHONY: deps
deps:  ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download

.PHONY: fmt
fmt:  ## Format Go code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

.PHONY: vet
vet:  ## Vet Go code
	@echo "Vetting code..."
	$(GOCMD) vet ./...

.PHONY: check
check: fmt vet lint test 
	@echo "All checks completed successfully!"
