package goal

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Go Concept - Background Worker Architecture:
// In Go, long-running background tasks are executed concurrently using lightweight goroutines.
// Instead of heavy OS threads, Go manages thousands of goroutines with minimal memory overhead (~2KB stack).
// We use a Producer-Consumer pattern where one producer queries for work and multiple worker goroutines
// process jobs concurrently via a thread-safe Go channel.

// StartWorker initializes and starts the month-end rollover background worker.
// Go Concepts used:
// 1. goroutines: concurrent execution with `go func()`
// 2. channels: safe inter-goroutine communication without explicit locks
// 3. sync.WaitGroup: tracks active goroutines to ensure zero leaked tasks during shutdown
// 4. time.Ticker: periodic recurring event triggers
// 5. context cancellation: gracefully halts workers when OS signals arrival
func StartWorker(ctx context.Context, svc *Service, interval time.Duration, wg *sync.WaitGroup) {
	// Go Concept - Buffered Channels:
	// A buffered channel `make(chan Type, capacity)` allows sending goroutines to enqueue
	// items without blocking until the buffer is full. Here 50 jobs can queue up cleanly.
	jobs := make(chan RolloverJob, 50)

	// Go Concept - Worker Goroutines Pool:
	// Spin up 3 concurrent worker goroutines reading from the shared jobs channel.
	// Each worker processes one month-close job at a time.
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(workerID int) {
			// Go Concept - sync.WaitGroup.Done:
			// defer wg.Done() decrements the WaitGroup counter when this goroutine returns.
			defer wg.Done()

			// Go Concept - Channel Range Loop:
			// `for job := range jobs` reads from the channel until the channel is closed by the producer.
			// When the channel closes and drains, the loop automatically terminates.
			for job := range jobs {
				res, err := svc.CloseMonth(ctx, job.UserID, job.Month)
				if err != nil {
					// Structured logging via log/slog ensures we record failures without crashing the process
					slog.Error("rollover worker job failed",
						"worker_id", workerID,
						"user_id", job.UserID,
						"month", job.Month.Format("2006-01"),
						"error", err,
					)
					continue
				}

				slog.Info("rollover worker job completed",
					"worker_id", workerID,
					"user_id", job.UserID,
					"month", job.Month.Format("2006-01"),
					"already_closed", res.AlreadyClosed,
					"moved_paise", res.MovedPaise,
					"goals_funded", res.GoalsFunded,
				)
			}
		}(i)
	}

	// Producer goroutine: scans for unclosed months and dispatches jobs to workers
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Closing the channel signals the 3 worker goroutines that no further jobs will arrive
		defer close(jobs)

		// Go Concept - time.Ticker:
		// time.Ticker periodically delivers ticks over a channel (ticker.C) at regular intervals.
		// For viva demos, set WORKER_INTERVAL=10s in .env; in production, 1h is typical.
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// scanAndEnqueue queries unclosed months for the previous calendar month and enqueues them
		scanAndEnqueue := func() {
			prevMonth := getPreviousMonth(time.Now())
			pendingJobs, err := svc.GetUnclosedPreviousMonths(ctx, prevMonth)
			if err != nil {
				slog.Error("worker producer failed to scan unclosed months", "error", err)
				return
			}

			for _, job := range pendingJobs {
				// Go Concept - Context Cancellation & Select:
				// `select` waits on multiple channel operations simultaneously.
				// If ctx.Done() signals during enqueuing, we abort immediately instead of hanging.
				select {
				case <-ctx.Done():
					return
				case jobs <- job:
				}
			}
		}

		// Execute one scan immediately upon server startup
		scanAndEnqueue()

		// Event loop listening for ticker pulses or shutdown signals
		for {
			select {
			// Go Concept - Context Cancellation:
			// When ctx is cancelled by SIGINT/SIGTERM, this case executes, exiting the goroutine cleanly.
			case <-ctx.Done():
				slog.Info("background rollover worker producer received shutdown signal")
				return
			case <-ticker.C:
				scanAndEnqueue()
			}
		}
	}()
}

// getPreviousMonth computes the 1st of the previous calendar month in UTC.
func getPreviousMonth(t time.Time) time.Time {
	utc := t.UTC()
	currentMonthStart := time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC)
	return currentMonthStart.AddDate(0, -1, 0)
}

