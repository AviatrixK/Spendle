# Spendle — Phase II Demonstration Script

> **Time Target:** 6–8 minutes  
> **Audience:** Project Evaluators & Course Faculty  
> **Key Objective:** Demonstrate working full-stack implementation with a focus on Go concurrency, database transactions, clean architecture, and responsive UX.

---

## 0. Pre-Flight Setup (Before Examiners Arrive)

1. **PostgreSQL running:**
   ```powershell
   docker start spendle-postgres
   ```
2. **Backend `.env` configured with short demo worker interval:**
   Ensure `backend/.env` has:
   ```env
   WORKER_INTERVAL=10s
   ```
3. **Start Go Backend in Terminal 1:**
   ```powershell
   cd d:\VsCode\Spendle\backend
   go run ./cmd/api
   ```
4. **Start React Frontend in Terminal 2:**
   ```powershell
   cd d:\VsCode\Spendle\frontend
   npm run dev
   ```
5. **Browser open:** Navigate to `http://localhost:5173` (on the `/login` screen).
6. **Seed command NOT yet run** (we will execute it live to show the worker waking up).

---

## Live Demo Flow

### Step 1: 30-Second Intro — Problem & Concept (0:00 - 0:30)
- **What to say:**  
  *"Most budget apps require tedious manual logging, and unspent budget at the end of the month is simply forgotten instead of saved. Spendle solves this by combining quick one-tap logging with an automated Go background worker that sweeps leftover budget directly into prioritized savings goals."*
- **What it shows:** Clear problem definition, domain understanding, and product differentiation.

---

### Step 2: Live Seed & Goroutine Worker Showcase (0:30 - 1:30)
- **What to type:**  
  Open Terminal 3 in PowerShell and run:
  ```powershell
  cd d:\VsCode\Spendle\backend
  go run ./cmd/seed
  ```
- **What to point at:**  
  Immediately switch your screen to **Terminal 1** (the running backend). Within 10 seconds, point at the structured log:
  ```
  INFO background rollover worker processed month close worker_id=1 user_id=... month=2026-10 already_closed=false moved_paise=750200 goals_funded=1
  ```
- **What to say:**  
  *"The seed script just populated 3 months of data, and our background Go worker, running on a 10-second ticker, immediately detected that last month had unspent budget and swept ₹7,502 into our top savings goal without any user interaction."*
- **Go Concept demonstrated:** Concurrent goroutines, `time.Ticker`, buffered channels (`chan RolloverJob`), and `slog` structured logging.

---

### Step 3: Login & Dashboard Overview (1:30 - 2:30)
- **What to click:**  
  In the browser at `http://localhost:5173/login`, log in with:
  - **Email:** `demo@spendle.app`
  - **Password:** `demo12345`
- **What to point out:**
  1. **Balance Banner:** Points out the net monthly position in green ("In profit").
  2. **Budgets List:** Scroll through categories. Point out the **Online Food** row in the previous month (`‹` button): notice it is visibly highlighted in coral and says `"₹450 over"` because spending exceeded the ₹1,500 limit.
  3. **Goals Preview:** Shows the top goals already reflecting progress from the worker sweep.
- **What to say:**  
  *"Our dashboard loads its balance summary, budget progress, and savings goals concurrently using Go's `errgroup`. The UI instantly reflects overspent categories with a coral visual indicator and capped progress bar."*
- **Go Concept demonstrated:** Concurrent fan-out with `errgroup.WithContext` and context timeout.

---

### Step 4: One-Tap Quick Add (2:30 - 3:45)
- **What to click:**  
  1. Tap the bottom floating **`+ Expense`** button.
  2. Notice the spring bottom sheet slides up with the amount input auto-focused.
  3. Type `250`, tap the `Food` category chip, tap `UPI`, and click **`Save Expense`**.
  4. Notice the bottom sheet slides down, the Food budget immediately increments, and the transaction appears at the top of Recent Transactions.
  5. Tap **`+ Income`**, type `5000`, tap `Cash`, type note `Part-time project`, and tap **`Save Income`**.
- **What to say:**  
  *"Spendle emphasizes one-tap logging for speed: users tap the floating button, type the amount, tap a category chip, and save in under three seconds."*
- **Go & Frontend Concept demonstrated:** Framer Motion spring physics, TanStack Query optimistic invalidation, and strict paise (`int64`) math.

---

### Step 5: Savings Goals Management (3:45 - 4:45)
- **What to click:**  
  1. Click **`Goals`** in the top navigation bar.
  2. Show the 3 goals ranked by priority (`#1 Bike`, `#2 MacBook Air`, `#3 Goa trip`).
  3. Point out the computed fields:
     - Bike has a deadline: shows `₹14,249/mo needed`.
     - MacBook Air has a monthly contribution: shows `ETA: Nov 2027`.
  4. Click **`Edit`** on the MacBook Air goal, change its Priority rank to `1`, and click **`Update Goal`**.
  5. Notice the list instantly re-sorts by priority.
- **What to say:**  
  *"Savings goals support both deadline-driven targets and monthly installment targets. Computed fields like monthly needed paise and ETA months are calculated as pure Go functions without database overhead."*
- **Go Concept demonstrated:** Pure calculation functions in `calc.go`, priority sorting, and validation.

---

### Step 6: Manual Month Close & Idempotency Demo (4:45 - 5:30)
- **What to click:**  
  1. In the **Month-End Rollover Sweep** panel at the top of the Goals page, select two months ago (e.g. `2026-09`).
  2. Click **`Close Month`**.
  3. Point to the green result banner: `"Rollover complete! Moved ₹8,053 into Bike."`
  4. Click **`Close Month`** again for the exact same month.
  5. Point to the result banner: `"Month 2026-09 was already closed. No duplicate rollover applied."`
- **What to say:**  
  *"The rollover uses an idempotent SQL transaction with `ON CONFLICT DO NOTHING`. If a user or worker triggers a sweep twice, the database safely rejects duplicate allocations."*
- **Go Concept demonstrated:** ACID database transactions (`tx.Begin`) and idempotent database operations.

---

### Step 7: Code Walkthrough & Live Unit Tests (5:30 - 7:00)
- **Files to show in VS Code:**
  1. [`backend/internal/goal/worker.go`](file:///d:/VsCode/Spendle/backend/internal/goal/worker.go): Show the producer goroutine, `time.Ticker`, buffered `chan RolloverJob`, and the 3 worker goroutines.
  2. [`backend/internal/dashboard/service.go`](file:///d:/VsCode/Spendle/backend/internal/dashboard/service.go): Show `errgroup.WithContext` launching 3 goroutines concurrently with a 3-second `context.WithTimeout`.
  3. [`backend/internal/goal/rollover.go`](file:///d:/VsCode/Spendle/backend/internal/goal/rollover.go): Show the pure `Allocate` function.
- **Live Terminal Command to Run:**  
  In PowerShell, run the live unit tests:
  ```powershell
  cd d:\VsCode\Spendle\backend
  go test -v ./internal/goal/... ./internal/dashboard/...
  ```
  *(Point out that all 4 required unit tests pass cleanly in under 5 milliseconds with zero live database dependency)*
- **What to say:**  
  *"We follow clean architecture: handler -> service -> repository with consumer-defined interfaces. Business logic like `Allocate` and `CalculateStatus` are pure functions tested with standard Go unit tests."*

---

### Step 8: Wrap-up & Phase III Roadmap (7:00 - 7:30)
- **What to say:**  
  *"In Phase II, we have delivered a fully working Go backend and React frontend adhering to strict paise precision, cookie auth, and native Go concurrency. For Phase III, we plan to implement automated recurring bill detection, multi-user split expenses, and containerized cloud deployment on Railway."*

---

## 🚨 "If Something Goes Wrong" Emergency Playbook

| Problem During Demo | Cause | Instant 10-Second Fix |
| :--- | :--- | :--- |
| **"database unavailable" or connection refused** | PostgreSQL Docker container stopped or crashed | Run in PowerShell: `docker start spendle-postgres` and refresh browser. |
| **"bind: address already in use :8080"** | A previous Go backend instance is still running | Run: `Stop-Process -Name main -Force` or find PID with `netstat -ano \| findstr :8080` and `taskkill /PID <PID> /F`. |
| **Login returns "authentication required" or loop** | Stale cookie from previous session | Open browser DevTools (`F12`) -> Application -> Cookies -> clear `spendle_token`, or open an Incognito window at `http://localhost:5173`. |

