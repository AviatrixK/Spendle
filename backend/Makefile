.PHONY: run build test tidy

# Run the API server in development mode
run:
	go run ./cmd/api

# Build the API binary into bin/
build:
	go build -o bin/api ./cmd/api

# Run all unit tests with verbose output
test:
	go test -v ./...

# Download and clean up dependencies in go.mod and go.sum
tidy:
	go mod tidy

