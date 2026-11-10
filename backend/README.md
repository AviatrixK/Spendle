# Spendle Backend

Backend API for **Spendle** — a budget tracker web application built for the "New Age Programming Languages Lab" course.

## Tech Stack
- **Language**: Go 1.22+
- **HTTP Router**: `github.com/go-chi/chi/v5`
- **Database Driver**: `github.com/jackc/pgx/v5` with plain SQL (NO ORM)
- **Migrations**: `golang-migrate`

---

## Directory Structure
```
backend/
├── cmd/
│   └── api/
│       └── main.go          # Application entrypoint & graceful shutdown
├── internal/
│   ├── config/              # Environment config loading and validation
│   ├── server/              # Server setup, router, route definitions & tests
│   ├── httpx/               # HTTP response helpers (JSON, error envelopes, decoding)
│   ├── middleware/          # Request logger, panic recoverer, CORS middleware
│   ├── auth/                # Authentication package placeholder (doc.go)
│   ├── budget/              # Budgeting package placeholder (doc.go)
│   ├── transaction/         # Income/expense transactions placeholder (doc.go)
│   ├── goal/                # Savings goals placeholder (doc.go)
│   └── dashboard/           # Metrics & summary placeholder (doc.go)
├── .env.example             # Example environment configuration
├── Makefile                 # Make targets for development and testing
└── README.md                # Backend documentation
```

---

## Getting Started

### 1. Prerequisites
- Go 1.22 or higher installed ([golang.org](https://go.dev/))
- Git

### 2. Configure Environment Variables
Copy the example environment file:
```bash
cp .env.example .env
```
Default values:
- `PORT=8080`
- `ENV=development`
- `FRONTEND_ORIGIN=http://localhost:5173`
- `DATABASE_URL=postgres://postgres:postgres@localhost:5432/spendle?sslmode=disable`
- `JWT_SECRET=spendle-super-secret-dev-key-change-in-production`

---

## Running the Server

### Using `go run`
From the `backend` directory:
```bash
go run ./cmd/api
```

### Using `make`
```bash
make run
```

When started, you will see structured log output:
```
time=... level=INFO msg="starting spendle api server" addr=:8080 env=development
```

To stop the server, press `Ctrl+C`. The server will catch the OS signal and perform a graceful shutdown within a 10-second timeout.

---

## Testing the `/ping` Endpoint

Once the server is running, test it using `curl` in another terminal:

### 1. Root Ping (`/ping`)
```bash
curl -i http://localhost:8080/ping
```
Expected response:
```http
HTTP/1.1 200 OK
Content-Type: application/json
Date: ...
Content-Length: 33

{"app":"spendle","status":"ok"}
```

### 2. Versioned Ping (`/api/v1/ping`)
```bash
curl -i http://localhost:8080/api/v1/ping
```
Expected response:
```http
HTTP/1.1 200 OK
Content-Type: application/json
Date: ...
Content-Length: 33

{"app":"spendle","status":"ok"}
```

---

## Running Unit Tests

Run the test suite using `go test`:
```bash
go test -v ./...
```

Or using `make`:
```bash
make test
```

---

## Building the Binary

Compile a production-ready binary:
```bash
make build
```
The binary will be generated at `backend/bin/api` (or `backend/bin/api.exe` on Windows).

