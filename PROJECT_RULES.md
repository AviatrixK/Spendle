# Spendle - Project Rules & Guidelines

## 1. Project Context
**Spendle** is a modern personal budget and expense tracking web application. It is being built as a college capstone project for a "New Age Programming Languages Lab" course, with **Go** as the primary backend language.

### Core Domain & Rules
- **Monthly Budgeting**: Users can configure monthly budgets per category, including custom user-created categories.
- **One-Tap Logging**: Quick logging for both expenses and income across three primary payment modes: **UPI**, **Cash**, and **Card**.
- **Dashboard & Insights**: Displays per-category "spent vs left", dynamic visual progress, and overall monthly profit/loss summary.
- **Savings Goals**: Users can set savings goals prioritized by rank.
- **Automated Rollover**: At month end, unspent budget is automatically moved into savings goals based on priority rank.
- **Currency**: **INR only (₹)**.
- **Monetary Representation**: Money is **ALWAYS** stored and computed as **integer paise** (`int64` in Go, `BIGINT` in PostgreSQL). Never use floating-point types (`float32`, `float64`, `NUMERIC` with decimals) for money to avoid floating-point inaccuracies. Example: `₹12.50` = `1250` paise.

---

## 2. Architecture & Patterns
- **Layered Architecture**: Strictly follow the unidirectional flow:
  ```
  HTTP Request -> Handler -> Service -> Repository -> Database
  ```
- **Interface-Driven Dependency Injection**: Services must depend on repository **interfaces** (defined alongside or where consumed), not concrete structs. This guarantees that all business logic in services can be unit-tested using mocks or stubs without requiring a live database.
- **Simplicity & Idiomatic Go**: 
  - Keep functions small and focused.
  - Return errors explicitly (`val, err := ...`); do not use `panic` for control flow.
  - Explain new or advanced Go concepts in comments (e.g., goroutines, channels, interfaces, defer, context, slog).

---

## 3. Technology Stack & Hard Constraints
*(Do NOT deviate from these constraints or introduce unlisted dependencies without explicit approval)*

### Backend
- **Language**: Go 1.22+
- **HTTP Router**: `github.com/go-chi/chi/v5`
- **Database Driver**: `github.com/jackc/pgx/v5` with plain SQL.
  - **NO ORM** (no GORM, no Ent).
  - **NO sqlc**.
- **Database Migrations**: `golang-migrate` (`github.com/golang-migrate/migrate/v4`).
- **Standard Library First**: Prefer Go's built-in libraries (`log/slog`, `net/http`, `context`, `encoding/json`, `time`, `crypto/rand`).

### Authentication (Future Task)
- Password hashing with `golang.org/x/crypto/bcrypt`.
- JWT-based authentication stored exclusively in secure, `httpOnly`, `SameSite` cookies (no local storage tokens).

### Frontend (Future Task)
- React + Vite
- Tailwind CSS
- Framer Motion
- Chart.js

---

## 4. Code & Workflow Constraints
- **Zero Extraneous Dependencies**: Do NOT add any third-party library that is not explicitly listed in the hard constraints without asking first.
- **Scope Discipline**: Do NOT implement features beyond what the active task asks for.
- **Comments for Learning**: Since the developer is transitioning from Node/Express to Go, include concise 1-2 line explanatory comments above Go-specific idioms and paradigms.

