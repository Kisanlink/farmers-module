# Variables
BINARY_NAME=build
SRC_FILE=main.go

# Targets

# Run the application with Air
air:
	air

# Run the application normally
run:
	go run $(SRC_FILE)

# Build the application binary
build:
	go build -o $(BINARY_NAME) $(SRC_FILE)

# Run tests
test:
	go test ./... -v

# Clean up generated files (Linux/macOS)
clean-linux:
	@rm -f $(BINARY_NAME)

# Clean up generated files (Windows)
clean-windows:
	@if exist $(BINARY_NAME) del $(BINARY_NAME)

# Format the Go code
fmt:
	go fmt ./...

# Install dependencies
deps:
	go mod tidy


