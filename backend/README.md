# Spendle Backend API 🛠️

Backend service for **Spendle** — a personal budget tracker and automated savings platform built for the **"New Age Programming Languages Lab"** course.

---

## 📋 Technology Stack

- **Language**: Go 1.22+
- **HTTP Router**: `github.com/go-chi/chi/v5`
- **Database Driver**: `github.com/jackc/pgx/v5` with plain SQL (*No ORM, No sqlc*)
- **Migrations**: `golang-migrate`
- **Logging**: Standard library `log/slog` (structured logging)

---

## 🏗️ Architecture & Design

Spendle's backend implements a strict **Handler -> Service -> Repository** layered architecture:

```
Request ──> [Middleware Chain] ──> [Handler] ──> [Service (Business Logic)] ──> [Repository (SQL)] ──> [PostgreSQL]
```

### Key Architectural Tenets
1. **Repository Interfaces**: Services depend exclusively on repository interfaces rather than concrete database connections. This ensures services can be tested with mock implementations without requiring a live database.
2. **Strict Integer Paise Storage**: Currency is stored as `int64` paise (`BIGINT` in Postgres). Never use `float` types for money to eliminate floating-point rounding errors. Example: `₹12.50` = `1250` paise.
3. **Graceful Shutdown**: The server manages OS termination signals (`SIGINT`, `SIGTERM`) through `signal.NotifyContext` with a 10-second deadline to ensure zero dropped in-flight requests.
4. **Standardized JSON Envelopes**: All responses use consistent shapes via the `httpx` package:
   - Success: Arbitrary payload (`{"status":"ok", ...}`)
   - Errors: Standard error envelope `{"error": "message"}`

---

## 📂 Directory Layout

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entrypoint & graceful shutdown orchestration
├── internal/
│   ├── config/                  # Environment variable loading, defaults & validation
│   ├── server/                  # Chi router, route registration & httptest suites
│   ├── httpx/                   # JSON response encoding, error envelopes & request decoders
│   ├── middleware/              # Slog request logger, panic recovery & CORS middleware
│   ├── auth/                    # User authentication, bcrypt & JWT (doc.go)
│   ├── budget/                  # Monthly category budgets & paise allocations (doc.go)
│   ├── transaction/             # Income & expense logging across UPI, Cash, Card (doc.go)
│   ├── goal/                    # Ranked savings goals & month-end rollover sweep (doc.go)
│   └── dashboard/               # Spent vs left metrics & profit/loss summaries (doc.go)
├── .env.example                 # Template for environment configuration
├── Makefile                     # Build, test, run, and dependency maintenance targets
└── README.md                    # Backend documentation
```

---

## ⚙️ Configuration & Environment Variables

Configuration is loaded from environment variables (or an optional `.env` file in the working directory) via `internal/config`:

| Variable | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `PORT` | Integer | `8080` | Port for the HTTP server to listen on |
| `ENV` | String | `development` | Environment mode (`development`, `production`, `test`) |
| `FRONTEND_ORIGIN` | String | `http://localhost:5173` | Allowed origin for CORS headers |
| `DATABASE_URL` | String | *Empty in dev* | PostgreSQL connection string (**required** in production) |
| `JWT_SECRET` | String | *Empty in dev* | Secret key for signing auth tokens (**required** in production) |

> [!IMPORTANT]
> When `ENV=production`, the application will fail fast on startup if `DATABASE_URL` or `JWT_SECRET` is missing.

---

## 🚀 Getting Started

### 1. Prerequisites
- **Go 1.22+** installed: [Verify with `go version`]
- **cURL** or an API client (e.g. Postman, Thunder Client)

### 2. Configure Environment
Create a `.env` file in `backend/` from the template:
```bash
cp .env.example .env
```

### 3. Run the Server
Using standard Go command:
```bash
go run ./cmd/api
```

Or using `make`:
```bash
make run
```

Log output will confirm the listener is active:
```
time=2026-11-10T15:58:15.000+05:30 level=INFO msg="starting spendle api server" addr=:8080 env=development
```

To stop the server safely, press `Ctrl+C`. The server will log its graceful shutdown sequence and exit cleanly.

---

## 📡 API Endpoints

### Health Check

#### 1. Root Ping: `GET /ping`
```bash
curl -i http://localhost:8080/ping
```

**Response:** `200 OK`
```json
{
  "app": "spendle",
  "status": "ok"
}
```

#### 2. Versioned Ping: `GET /api/v1/ping`
```bash
curl -i http://localhost:8080/api/v1/ping
```

**Response:** `200 OK`
```json
{
  "app": "spendle",
  "status": "ok"
}
```

---

## 🧪 Testing

The test suite utilizes Go's standard `testing` package and `net/http/httptest` to test HTTP routers and handlers in-memory without opening network ports.

Run all tests:
```bash
go test -v ./...
```

Or via `make`:
```bash
make test
```

---

## 🔨 Build & Maintenance (Makefile)

| Command | Action |
| :--- | :--- |
| `make run` | Starts the API server using `go run ./cmd/api` |
| `make build` | Compiles an optimized binary into `bin/api` |
| `make test` | Executes all package tests with verbose logging |
| `make tidy` | Resolves, downloads, and removes unused dependencies |

---

## 💡 Go Idioms Reference for Beginners

If you are coming from Node/Express, here is how key concepts translate to this codebase:

- **Error Handling**: Instead of `try/catch` or `next(err)`, Go functions return `(result, error)`. Always check `if err != nil`.
- **Structs & Interfaces**: Structs define data shapes. Interfaces define method signatures. Types satisfy interfaces implicitly (no `implements` keyword).
- **Defer**: `defer cleanup()` schedules code to execute immediately before the surrounding function exits (ideal for closing files or recovering from panics).
- **Goroutines**: `go srv.Start()` spawns a lightweight concurrent thread managed by the Go runtime.
- **Context**: `context.Context` transmits cancellation signals, timeouts, and deadlines across API handlers and background processes.
