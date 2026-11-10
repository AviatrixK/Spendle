# Spendle Backend API 🛠️

Backend service for **Spendle** — a personal budget tracker and automated savings platform built for the **"New Age Programming Languages Lab"** course.

---

## 📋 Technology Stack

- **Language**: Go 1.22+
- **HTTP Router**: `github.com/go-chi/chi/v5`
- **Database Driver**: `github.com/jackc/pgx/v5` (`pgxpool`) with plain SQL (*No ORM, No sqlc*)
- **Migrations**: `github.com/golang-migrate/migrate/v4` (embedded using `//go:embed` and `iofs`)
- **Logging**: Standard library `log/slog` (structured logging)

---

## 🏗️ Architecture & Design

Spendle's backend implements a strict **Handler -> Service -> Repository** layered architecture:

```
Request ──> [Middleware Chain] ──> [Handler] ──> [Service (Business Logic)] ──> [Repository (SQL)] ──> [PostgreSQL]
```

### Key Architectural Tenets
1. **Repository Interfaces**: Services depend exclusively on repository interfaces rather than concrete database connections. This ensures services can be tested with mock implementations without requiring a live database.
2. **Strict Integer Paise Storage**: Currency is stored strictly as `int64` paise (`BIGINT` in PostgreSQL). Never use `float` types for money to eliminate floating-point rounding errors. Example: `₹12.50` = `1250` paise.
3. **Database Connection Pooling**: PostgreSQL connections are managed via a tuned `pgxpool.Pool` (with connection limits, idle timeouts, and health checks).
4. **Embedded Migrations**: All SQL schema migrations reside in `internal/db/migrations/` and are compiled into the binary via `//go:embed`, running automatically on startup.
5. **Graceful Shutdown**: The server manages OS termination signals (`SIGINT`, `SIGTERM`) through `signal.NotifyContext` with a 10-second deadline to ensure active requests finish and database pools close cleanly.
6. **Standardized JSON Envelopes**: All responses use consistent shapes via the `httpx` package:
   - Success: Payload (`{"status":"ok", ...}`)
   - Errors: Standard error envelope `{"error": "message"}`

---

## 📂 Directory Layout

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entrypoint & lifecycle orchestration
├── internal/
│   ├── config/                  # Environment variable loading, defaults & validation
│   ├── db/                      # Connection pool initialization & embedded migrations
│   │   ├── db.go                # pgxpool setup and health ping
│   │   ├── migrate.go           # golang-migrate runner with //go:embed and iofs
│   │   └── migrations/          # Versioned SQL migration files (.up.sql / .down.sql)
│   │       ├── 000001_init.up.sql
│   │       └── 000001_init.down.sql
│   ├── server/                  # Chi router, health checks & httptest suites
│   ├── httpx/                   # JSON response encoding, error envelopes & request decoders
│   ├── middleware/              # Slog request logger, panic recovery & CORS middleware
│   ├── auth/                    # User authentication, bcrypt & JWT (doc.go)
│   ├── budget/                  # Monthly category budgets & paise allocations (doc.go)
│   ├── transaction/             # Income & expense logging across UPI, Cash, Card (doc.go)
│   ├── goal/                    # Ranked savings goals & month-end rollover sweep (doc.go)
│   └── dashboard/               # Spent vs left metrics & profit/loss summaries (doc.go)
├── .env.example                 # Template for environment configuration
├── Makefile                     # Build, test, run, and database lifecycle targets
└── README.md                    # Backend documentation
```

---

## ⚙️ Configuration & Environment Variables

Configuration is loaded from environment variables (or an optional `.env` file in the working directory) via `internal/config`:

| Variable | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `PORT` | Integer | `8080` | Port for the HTTP server to listen on |
| `ENV` | String | `development` | Environment mode (`development`, `production`, `test`) |
| `DATABASE_URL` | String | *(Required)* | PostgreSQL connection URL (e.g. `postgres://spendle:spendle@localhost:5432/spendle_db?sslmode=disable`) |
| `JWT_SECRET` | String | *Empty in dev* | Secret key for signing auth tokens (**required** in production) |
| `FRONTEND_ORIGIN` | String | `http://localhost:5173` | Allowed origin for CORS headers |

> [!IMPORTANT]
> `DATABASE_URL` is required across all environments. If missing or empty, startup will fail immediately with a descriptive error.

---

## 🚀 Getting Started

### 1. Prerequisites
- **Go 1.22+** installed: [Verify with `go version`]
- **Docker & Docker Compose** installed: [Verify with `docker compose version`]
- **cURL** or an API client (e.g. Postman, Thunder Client)

### 2. Configure Environment
Create a `.env` file in `backend/` from the template:
```bash
cp .env.example .env
```

### 3. Start the PostgreSQL Database
Start the PostgreSQL 16 container in the background:
```bash
make db-up
```
*(Or run `docker compose up -d db` from the repository root).*

> [!NOTE]
> If host port `5432` is already in use by a local PostgreSQL service, update the port mapping in `docker-compose.yml` (e.g. `"5433:5432"`) and update `DATABASE_URL` in `backend/.env` to point to port `5433`.

### 4. Run the API Server
Start the server in development mode:
```bash
make run
```
*(Or run `go run ./cmd/api`).*

On startup, Spendle will automatically:
1. Run pending SQL migrations using the embedded `golang-migrate` runner.
2. Initialize and ping the PostgreSQL connection pool.
3. Start the HTTP listener on port `8080`.

Log output will confirm successful initialization:
```
time=... level=INFO msg="running database migrations..."
time=... level=INFO msg="database migrations completed successfully"
time=... level=INFO msg="connecting to database pool..."
time=... level=INFO msg="database connection pool initialized"
time=... level=INFO msg="starting spendle api server" addr=:8080 env=development
```

---

## 🔍 Database Inspection & Migrations

### Confirm Tables with `\dt`
To verify that all MVP schema tables were successfully created, connect to the database container using `psql`:

```bash
make psql
```
*(Or run: `docker compose exec -it db psql -U spendle -d spendle_db`)*

Inside the `psql` prompt, list all tables:
```sql
\dt
```

Expected table list:
```
               List of relations
 Schema |       Name        | Type  |  Owner  
--------+-------------------+-------+---------
 public | categories        | table | spendle
 public | goal_contributions| table | spendle
 public | goals             | table | spendle
 public | month_closings    | table | spendle
 public | monthly_budgets   | table | spendle
 public | schema_migrations | table | spendle
 public | transactions      | table | spendle
 public | users             | table | spendle
(8 rows)
```

To exit the `psql` shell, type `\q` and press Enter.

### How to Run Migrations Down

#### Option 1: Database Reset (Recommended for Local Dev)
To drop all tables, volumes, and rebuild the database from scratch:
```bash
make db-reset
```
When you next run `make run`, all `up` migrations will automatically reapply cleanly.

#### Option 2: Using the golang-migrate CLI
If you have the `migrate` CLI installed:
```bash
migrate -path internal/db/migrations -database "postgres://spendle:spendle@localhost:5432/spendle_db?sslmode=disable" down
```

---

## 📡 API Endpoints

### 1. Root Ping: `GET /ping`
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

### 2. Versioned Ping: `GET /api/v1/ping`
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

### 3. Database Health Check: `GET /api/v1/health`
Pings the database through the connection pool with a 2-second timeout:
```bash
curl -i http://localhost:8080/api/v1/health
```

**Success Response (Database Up):** `200 OK`
```json
{
  "database": "up",
  "status": "ok"
}
```

**Failure Response (Database Down/Unreachable):** `503 Service Unavailable`
```json
{
  "error": "database unavailable"
}
```

---

## 🧪 Testing

Run all unit and HTTP tests:
```bash
make test
```
*(Or run `go test -v ./...` inside `backend/`).*

---

## 🔨 Build & Maintenance (Makefile)

| Command | Action |
| :--- | :--- |
| `make run` | Starts the API server (runs migrations & initializes pool) |
| `make build` | Compiles an optimized binary into `bin/api` |
| `make test` | Executes all package tests with verbose logging |
| `make tidy` | Downloads and organizes dependencies in `go.mod` |
| `make db-up` | Starts the PostgreSQL container via Docker Compose |
| `make db-down` | Stops the PostgreSQL container |
| `make db-reset` | Wipes database volumes and restarts fresh container |
| `make psql` | Opens an interactive `psql` shell in the database container |

---

## 💡 Go Idioms Reference for Beginners

- **Connection Pool**: Reuses open TCP connections to PostgreSQL across requests instead of opening and closing sockets each time.
- **go:embed**: Direct compiler directive (`//go:embed`) to bundle SQL files directly into the compiled executable.
- **Context Timeouts**: `context.WithTimeout` aborts database operations if they take too long, preventing cascading hangs.
- **Error Values**: Errors are values returned explicitly `(result, err)` and checked with `if err != nil`.
- **Defer**: `defer pool.Close()` guarantees cleanup when the enclosing function terminates.
