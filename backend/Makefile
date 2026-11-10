.PHONY: run build test tidy db-up db-down db-reset psql

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

# Start the PostgreSQL database container in the background
db-up:
	docker compose up -d db

# Stop the PostgreSQL database container
db-down:
	docker compose down

# Reset the database by removing containers and volumes, then restarting fresh
db-reset:
	docker compose down -v
	docker compose up -d db

# Open an interactive psql shell inside the running database container
psql:
	docker compose exec -it db psql -U spendle -d spendle_db
