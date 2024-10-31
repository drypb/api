BINARY_NAME=api
BUILD_DIR=bin
BUILD_OPTIONS=CGO_ENABLED=0 GOOS=linux
GO_FILES=./cmd/api

.PHONY: all build run test clean

build: $(GO_FILES)
	@echo "Compiling project..."
	@$(BUILD_OPTIONS) go build -o $(BUILD_DIR)/$(BINARY_NAME) $(GO_FILES)
	@echo "Compilation concluded: $(BUILD_DIR)/$(BINARY_NAME)"

run: build
	@echo "Executing project..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

test:
	@echo "Executing tests..."
	@go test ./...

clean:
	@echo "Cleaning binaries..."
	@rm -rf $(BUILD_DIR)
	@echo "Binaries removed."

all: build

