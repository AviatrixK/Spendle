# Spendle 💰

> **Smart, disciplined personal budgeting and automated savings tracker.**  
> Built as a capstone project for the **New Age Programming Languages Lab** course, featuring **Go** as the primary backend language.

---

## 📌 Project Overview

**Spendle** helps individuals take control of their personal finances through structured monthly budgeting, instant transaction logging, and automated savings sweeps. Designed with strict financial accuracy and modern software architecture principles in mind.

### Key Capabilities
- **Category-Based Monthly Budgets**: Set dynamic spending targets for default and custom categories.
- **One-Tap Transaction Logging**: Quickly log income and expenses across three primary payment modes: **UPI**, **Cash**, and **Card**.
- **Visual Dashboard**: Real-time "spent vs. left" tracking per category, alongside overall monthly profit and loss summaries.
- **Ranked Savings Goals**: Set and prioritize personal savings targets (e.g., Emergency Fund #1, Tech Upgrade #2).
- **Automated Month-End Rollover**: Any unspent budget at the end of the month automatically sweeps into savings goals according to their priority rank.
- **Strict Monetary Precision (Zero Floats)**: All monetary figures are calculated and persisted strictly as **integer paise** (`int64` in Go, `BIGINT` in PostgreSQL). No floating-point rounding errors. Currency is INR (₹) exclusively.

---

## 🏛️ System Architecture

Spendle follows a strict **Layered Architecture** with unidirectional dependency flow:

```
[ Client / Frontend ]
          │ (HTTP / JSON)
          ▼
   [ HTTP Handler ]      --> Parses requests, validates inputs, calls service layer
          │
          ▼
   [ Service Layer ]     --> Contains core business logic, depends on repository interfaces
          │
          ▼
 [ Repository Layer ]    --> Executes plain SQL queries against PostgreSQL using pgx
          │
          ▼
    [ PostgreSQL ]       --> Relational data store (amounts stored as BIGINT paise)
```

### Architectural Guiding Principles
1. **Interface-Driven Dependency Injection**: Services depend solely on repository *interfaces*, allowing business logic to be unit tested in isolation without a live database.
2. **Standard Library First**: Built with Go's robust standard packages (`net/http`, `log/slog`, `context`, `encoding/json`), minimizing external dependencies.
3. **No ORM**: Plain, explicit SQL via `pgx/v5` for predictable database performance and full query visibility.
4. **Graceful Lifecycle Management**: Clean server startup and graceful termination handling OS interrupts (`SIGINT`, `SIGTERM`) with context deadlines.

---

## 🛠️ Technology Stack

| Layer | Technology | Details |
| :--- | :--- | :--- |
| **Backend Language** | **Go 1.22+** | Idiomatic, strongly-typed, concurrent backend |
| **HTTP Router** | **Chi v5** (`github.com/go-chi/chi/v5`) | Lightweight, idiomatic standard `net/http` compatible routing |
| **Database Driver** | **pgx v5** (`github.com/jackc/pgx/v5`) | High-performance PostgreSQL driver (Plain SQL, NO ORM) |
| **Migrations** | **golang-migrate** | Reproducible, versioned database schema migrations |
| **Logging** | **log/slog** | Go standard library structured logging (JSON in prod, text in dev) |
| **Frontend** *(Upcoming)* | **React + Vite** | Fast, responsive modern web client |
| **Styling & UI** *(Upcoming)* | **Tailwind CSS + Framer Motion** | Utility-first styling with smooth interactive animations |
| **Charts** *(Upcoming)* | **Chart.js** | Visual spending analytics and progress meters |

---

## 📂 Repository Structure

```
Spendle/
├── .gitignore               # Root git ignore (Go binaries, .env, node_modules)
├── PROJECT_RULES.md         # Hard project rules, domain guidelines & architectural constraints
├── README.md                # Project-wide documentation
│
├── backend/                 # Go backend service
│   ├── cmd/
│   │   └── api/
│   │       └── main.go      # Entrypoint & graceful shutdown orchestration
│   ├── internal/
│   │   ├── auth/            # Authentication, JWT, and session security
│   │   ├── budget/          # Category budgets & paise allocations
│   │   ├── config/          # Environment variable loading & validation
│   │   ├── dashboard/       # Aggregated financial analytics & metrics
│   │   ├── goal/            # Savings goals & priority sweep engine
│   │   ├── httpx/           # Reusable HTTP JSON helpers & error envelopes
│   │   ├── middleware/      # Slog logger, panic recovery, CORS middleware
│   │   ├── server/          # Chi router setup, routes & httptest suites
│   │   └── transaction/     # Income/expense transactions (UPI, Cash, Card)
│   ├── .env.example         # Template configuration for local environments
│   ├── Makefile             # Development automation (run, build, test, tidy)
│   └── README.md            # Backend-specific documentation & setup guide
│
└── Frontend/                # React client application (Upcoming)
```

---

## 🚀 Quickstart Guide

### 1. Prerequisites
- **Go**: 1.22 or higher ([Download Go](https://go.dev/dl/))
- **Git**
- **PostgreSQL**: 15+ (for subsequent database tasks)

### 2. Setup the Backend
Navigate to the `backend/` directory:
```bash
cd backend
```

Copy the example environment file:
```bash
cp .env.example .env
```

Verify or update the configuration in `.env`:
```env
PORT=8080
ENV=development
FRONTEND_ORIGIN=http://localhost:5173
DATABASE_URL=postgres://postgres:postgres@localhost:5432/spendle?sslmode=disable
JWT_SECRET=super-secret-development-key-change-in-production
```

### 3. Run the Backend Server
```bash
go run ./cmd/api
```
*(Or use `make run`)*

You should see structured log output confirming server startup:
```
time=... level=INFO msg="starting spendle api server" addr=:8080 env=development
```

### 4. Verify API Endpoints
Open a new terminal window and run:

```bash
# Test root health check
curl -i http://localhost:8080/ping

# Test versioned health check
curl -i http://localhost:8080/api/v1/ping
```

Expected JSON response:
```json
{
  "app": "spendle",
  "status": "ok"
}
```

### 5. Run Unit Tests
```bash
go test -v ./...
```
*(Or use `make test`)*

---

## ⚙️ Development Commands (Makefile)

From within the `backend/` directory:

| Command | Action |
| :--- | :--- |
| `make run` | Starts the HTTP API server in development mode |
| `make build` | Compiles the production binary into `backend/bin/api` |
| `make test` | Executes all package unit tests with verbose output |
| `make tidy` | Resolves, downloads, and cleans up `go.mod` and `go.sum` |

---

## 📜 Coding Conventions & Guidelines

- **Zero Float Rule**: Never use floats for financial math. Store as `int64` paise (e.g., ₹12.50 = `1250` paise).
- **Explicit Error Handling**: Always return `(result, error)`. Never panic for expected runtime issues.
- **Educational Comments**: Key Go idioms (`defer`, `goroutines`, `channels`, `context`, `interfaces`) are documented inline for reference.
- Detailed rules are codified in [`PROJECT_RULES.md`](file:///d:/VsCode/Spendle/PROJECT_RULES.md).

