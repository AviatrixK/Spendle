# Spendle — Capstone Viva Study Notes

> **Target:** New Age Programming Languages Lab Examination  
> **Primary Technology:** Go 1.22+ (Backend) & PostgreSQL 16  
> **Key Architectural Paradigm:** Layered Architecture (`Handler -> Service -> Repository`), Plain SQL with `pgx`, and Concurrency Primitives.

---

## 1. Goroutines & Channels (The Background Worker)

### Plain Definition
A goroutine is a lightweight thread of execution managed by the Go runtime, while a channel is a typed conduit through which goroutines synchronize and communicate data safely without explicit locks.

### Project Location
- **File:** [`backend/internal/goal/worker.go`](file:///d:/VsCode/Spendle/backend/internal/goal/worker.go)
- **Functions:** `StartWorker()`, `scanAndEnqueue()`

### Why Chosen Over Alternatives
Instead of spinning up heavy operating system threads (which allocate ~1MB stack memory each) or relying on external task queues like Redis/Celery, Go goroutines start with just 2KB of stack space. A producer-consumer channel (`jobs := make(chan RolloverJob, 50)`) allows 3 worker goroutines to process rollover sweeps concurrently with zero shared-memory race conditions.

### Likely Examiner Questions & Model Answers
- **Q1: Why did you use a buffered channel instead of an unbuffered channel in the worker?**  
  *Answer:* An unbuffered channel blocks the sender until a receiver is ready. A buffered channel of size 50 allows the producer to enqueue multiple unclosed months in one quick database scan without pausing on every single item, smoothing out database I/O bursts.
- **Q2: What prevents worker goroutines from leaking when the server shuts down?**  
  *Answer:* When `ctx.Done()` signals shutdown, the producer exits its loop, invokes `defer close(jobs)`, and calls `defer wg.Done()`. The workers range over the channel until it drains and closes, then exit cleanly. `wg.Wait()` in `main.go` blocks until all worker goroutines have returned.

---

## 2. Context Cancellation & Graceful Shutdown

### Plain Definition
A Go `context.Context` carries deadlines, cancellation signals, and request-scoped values across API boundaries and concurrent goroutines.

### Project Location
- **Files:** [`backend/cmd/api/main.go`](file:///d:/VsCode/Spendle/backend/cmd/api/main.go), [`backend/internal/server/server.go`](file:///d:/VsCode/Spendle/backend/internal/server/server.go)
- **Functions:** `signal.NotifyContext()`, `server.Shutdown(shutdownCtx)`

### Why Chosen Over Alternatives
Abruptly terminating an application with `os.Exit(0)` or `SIGKILL` cuts off active client HTTP requests mid-flight and corrupts in-progress database writes. `signal.NotifyContext` catches OS termination signals and allows in-flight HTTP requests and background worker sweeps up to 10 seconds to finish cleanly before terminating the database pool.

### Likely Examiner Questions & Model Answers
- **Q1: How does the main goroutine know an OS interrupt (Ctrl+C) occurred?**  
  *Answer:* `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` returns a context whose `Done()` channel closes when the OS delivers the signal. The main thread blocks on `<-ctx.Done()`, waking up immediately to initiate graceful cleanup.
- **Q2: Why do you pass a separate timeout context to `server.Shutdown()`?**  
  *Answer:* The root signal context is already cancelled when the interrupt arrives. Passing `context.WithTimeout(context.Background(), 10*time.Second)` creates an active 10-second grace window for TCP connections to finish writing their responses.

---

## 3. Concurrency with `errgroup` (The Dashboard Aggregator)

### Plain Definition
`errgroup.Group` coordinates a collection of goroutines working on subtasks of a single operation, synchronizing their completion and propagating the first error encountered.

### Project Location
- **File:** [`backend/internal/dashboard/service.go`](file:///d:/VsCode/Spendle/backend/internal/dashboard/service.go)
- **Function:** `Service.GetDashboard()`

### Why Chosen Over Alternatives
If loaded sequentially, the dashboard would incur the latency sum: `latency(summary) + latency(budgets) + latency(goals)`. Using `errgroup.WithContext`, all three database operations execute in parallel, reducing response time to `max(latency1, latency2, latency3)`. If any query fails, the group context cancels the remaining subtasks immediately.

### Likely Examiner Questions & Model Answers
- **Q1: What happens if the `GetSummary` goroutine fails while `ListGoals` is still querying?**  
  *Answer:* `errgroup.WithContext(timeoutCtx)` derives a child context (`gCtx`). When `GetSummary` returns an error, `errgroup` cancels `gCtx`, notifying the database driver to cancel the remaining queries and returning the error to the HTTP handler.
- **Q2: Why use `errgroup` instead of a plain `sync.WaitGroup`?**  
  *Answer:* A plain `sync.WaitGroup` only synchronizes completion; it does not collect errors or propagate context cancellation. `errgroup` handles error collection, goroutine synchronization, and context cancellation in one clean abstraction.

---

## 4. Interfaces & The Repository Pattern

### Plain Definition
In Go, an interface defines a contract of behavior (method signatures) that any concrete struct implements implicitly without an explicit `implements` keyword.

### Project Location
- **Files:** [`backend/internal/goal/repository.go`](file:///d:/VsCode/Spendle/backend/internal/goal/repository.go), [`backend/internal/dashboard/service.go`](file:///d:/VsCode/Spendle/backend/internal/dashboard/service.go)
- **Interfaces:** `goal.Repository`, `dashboard.BudgetsReader`, `dashboard.SummaryReader`

### Why Chosen Over Alternatives
Directly coupling services to `*pgxpool.Pool` makes automated unit testing impossible without spinning up a live PostgreSQL instance. By depending strictly on repository interfaces, we can swap in in-memory mocks during unit testing. Furthermore, in `internal/dashboard`, interfaces are defined by the consumer, keeping domain packages loosely coupled.

### Likely Examiner Questions & Model Answers
- **Q1: What does "interfaces should be defined by the consumer" mean in Go?**  
  *Answer:* Instead of a package exporting wide interfaces containing every method it supports, the consuming package (`dashboard`) defines a small, targeted interface (`BudgetsReader`) containing only the single method it actually calls. This follows the Interface Segregation Principle.
- **Q2: Does Go require a struct to declare that it implements an interface?**  
  *Answer:* No. Go uses structural subtyping (duck typing). If a struct defines the matching method signatures with identical parameter and return types, the Go compiler automatically recognizes that it satisfies the interface.

---

## 5. Error Handling with Sentinel Errors & `errors.Is`

### Plain Definition
Sentinel errors are pre-allocated package-level error variables used to signal specific, expected domain error conditions that callers can inspect and handle programmatically.

### Project Location
- **Files:** [`backend/internal/goal/errors.go`](file:///d:/VsCode/Spendle/backend/internal/goal/errors.go), [`backend/internal/auth/errors.go`](file:///d:/VsCode/Spendle/backend/internal/auth/errors.go)
- **Variables:** `ErrGoalNotFound`, `ErrConflictingTimeline`, `ErrInvalidTarget`

### Why Chosen Over Alternatives
Comparing error strings (`err.Error() == "not found"`) is brittle and breaks whenever log phrasing changes. Sentinel errors wrapped with `fmt.Errorf("...: %w", err)` preserve the causal stack while enabling robust matching via `errors.Is(err, target)`.

### Likely Examiner Questions & Model Answers
- **Q1: Why should you use `errors.Is(err, ErrNotFound)` instead of `err == ErrNotFound`?**  
  *Answer:* Because Go errors are frequently wrapped using the `%w` verb in `fmt.Errorf`. Direct equality (`==`) checks fail on wrapped errors, whereas `errors.Is` recursively unwraps the error chain to inspect root causes.
- **Q2: What is the difference between `errors.Is` and `errors.As`?**  
  *Answer:* `errors.Is` checks for a specific sentinel error value. `errors.As` checks whether an error in the chain matches a specific error *type* and extracts its fields (such as inspecting `*pgconn.PgError` to read the PostgreSQL error code `"23505"`).

---

## 6. Authentication: Bcrypt Hashing & `httpOnly` JWT Cookies

### Plain Definition
Bcrypt is an adaptive, slow password hashing algorithm designed to resist brute-force attacks, and `httpOnly` cookies are browser storage attributes that prevent client-side JavaScript from accessing session tokens.

### Project Location
- **Files:** [`backend/internal/auth/service.go`](file:///d:/VsCode/Spendle/backend/internal/auth/service.go), [`backend/internal/auth/middleware.go`](file:///d:/VsCode/Spendle/backend/internal/auth/middleware.go)
- **Functions:** `Signup()`, `Login()`, `RequireAuth()`

### Why Chosen Over Alternatives
Storing JWTs in browser `localStorage` leaves tokens completely vulnerable to Cross-Site Scripting (XSS) extraction. By issuing the JWT inside a `Set-Cookie` header configured with `httpOnly: true`, `SameSite: Lax`, and `Path: /`, JavaScript can never read the raw token. Constant-time dummy bcrypt execution on missing accounts mitigates user enumeration timing attacks.

### Likely Examiner Questions & Model Answers
- **Q1: How does Spendle prevent timing attacks during user login?**  
  *Answer:* Bcrypt takes ~150ms. If an email does not exist, an early return would take <2ms, allowing attackers to measure response times to enumerate valid accounts. If the email is not found, Spendle executes bcrypt against a precomputed dummy hash, equalizing response times.
- **Q2: What is the maximum password length supported by bcrypt and how does Spendle handle it?**  
  *Answer:* The bcrypt algorithm truncates inputs longer than 72 bytes. Spendle explicitly validates `len(password) <= 72` in `auth.Service.Signup`, returning a 400 error rather than silently ignoring characters past byte 72.

---

## 7. Database Transactions & Idempotency (Month Close)

### Plain Definition
A database transaction guarantees ACID atomicity (all steps succeed together or all rollback), and idempotency guarantees that making the same request multiple times produces the exact same result without duplicate side-effects.

### Project Location
- **File:** [`backend/internal/goal/repository_pg.go`](file:///d:/VsCode/Spendle/backend/internal/goal/repository_pg.go)
- **Function:** `PGRepository.CloseMonth()`

### Why Chosen Over Alternatives
Sweeping leftover budget into goals involves four discrete mutations: checking closing status, calculating leftover, updating goal balances, and recording audit contributions. If any step fails without a transaction, money is partially credited or lost. Using `ON CONFLICT DO NOTHING` on `month_closings (user_id, month)` makes the operation safe to trigger repeatedly.

### Likely Examiner Questions & Model Answers
- **Q1: How does `CloseMonth` guarantee it will never sweep funds twice if the worker and user click close simultaneously?**  
  *Answer:* PostgreSQL enforces a primary key constraint on `month_closings (user_id, month)`. The transaction executes `INSERT ... ON CONFLICT DO NOTHING`. If `tag.RowsAffected() == 0`, another process already closed the month, so the transaction aborts and returns `AlreadyClosed: true`.
- **Q2: Why do you put `defer tx.Rollback(ctx)` immediately after `r.pool.Begin(ctx)`?**  
  *Answer:* In Go, `defer` guarantees execution on function return. If any query returns an error or panics, `tx.Rollback` cancels the transaction. If `tx.Commit()` succeeds at the end, `tx.Rollback` safely turns into a harmless no-op.

---

## 8. Monetary Standard: Strict Integer Paise (`int64`)

### Plain Definition
All monetary calculations and storage use integer units representing the smallest currency denomination (paise in INR, where ₹1.00 = 100 paise), completely prohibiting floating-point types (`float32`, `float64`).

### Project Location
- **Across the Entire Codebase:** `TargetPaise int64`, `SavedPaise int64`, `AmountPaise int64`
- **PostgreSQL Column Types:** `BIGINT NOT NULL`

### Why Chosen Over Alternatives
Standard floating-point numbers (`float64`) are based on IEEE 754 binary fractions, causing representation errors where `0.1 + 0.2 = 0.30000000000000004`. Over hundreds of financial transactions, floating-point rounding errors compound into discrepancies. Integer paise math guarantees 100% exact precision for addition, subtraction, and ceiling division.

### Likely Examiner Questions & Model Answers
- **Q1: How do you perform ceiling division for monthly targets in integer math without using floats?**  
  *Answer:* We use the integer formula `ceil(A / B) = (A + B - 1) / B`. For example, dividing 10000 paise by 3 months: `(10000 + 3 - 1) / 3 = 10002 / 3 = 3334 paise`, avoiding `math.Ceil` or float casting.
- **Q2: Why use `int64` instead of standard `int`?**  
  *Answer:* In Go, plain `int` is architecture-dependent (32-bit on 32-bit platforms, capping at ~₹2.14 crore). `int64` is guaranteed 64-bit across all platforms, safely handling values up to ₹92,233,720,368,547,758.07 without integer overflow.

---

## 9. Multi-Tenant Security: Mandatory User ID Scoping

### Plain Definition
Every database read, update, and deletion is explicitly scoped with `WHERE user_id = $1`, with the user ID securely extracted from the cryptographically verified JWT request context.

### Project Location
- **Middleware:** [`backend/internal/auth/middleware.go`](file:///d:/VsCode/Spendle/backend/internal/auth/middleware.go)
- **Repositories:** Every query in `internal/budget/repository_pg.go`, `internal/goal/repository_pg.go`, and `internal/transaction/repository_pg.go`

### Why Chosen Over Alternatives
Relying on user IDs supplied in request URLs or request bodies exposes an application to Insecure Direct Object References (IDOR), where an attacker modifies `?user_id=target-uuid` to read or delete another user's financial records. In Spendle, the client has zero control over user scoping.

### Likely Examiner Questions & Model Answers
- **Q1: How does the server prevent a user from deleting another user's transaction with `DELETE /transactions/{id}`?**  
  *Answer:* The handler extracts the authenticated user ID from context (`auth.UserIDFromContext(r.Context())`). The repository query executes `DELETE FROM transactions WHERE id = $1 AND user_id = $2`. If the transaction belongs to someone else, 0 rows are affected and the server returns `404 Not Found`.
- **Q2: Why is the context key in `internal/auth/middleware.go` defined as an unexported struct?**  
  *Answer:* Go context values can be overwritten if two packages use generic string keys like `"user_id"`. Declaring `type contextKey struct{}` creates a private, unexported type that no external package can forge, overwrite, or collide with.

