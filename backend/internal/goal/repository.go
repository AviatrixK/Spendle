package goal

import (
	"context"
	"time"
)

// Repository defines the persistence interface contract for savings goals and month-end sweeps.
// Go Concept - Interface Driven Architecture:
// In Go, interfaces declare method sets that concrete types implement implicitly.
// By having services depend on this interface rather than *pgxpool.Pool directly,
// we can easily substitute fake in-memory repositories in unit tests.
type Repository interface {
	// CreateGoal inserts a new savings goal for the user.
	CreateGoal(ctx context.Context, g Goal) (Goal, error)

	// GetGoalByID retrieves a single goal ensuring it belongs to the user.
	GetGoalByID(ctx context.Context, userID, goalID string) (Goal, error)

	// ListGoals returns all goals belonging to the user ordered by priority ASC, created_at ASC.
	ListGoals(ctx context.Context, userID string) ([]Goal, error)

	// UpdateGoal modifies title, target, timeline, priority, and status.
	UpdateGoal(ctx context.Context, g Goal) (Goal, error)

	// DeleteGoal removes a goal belonging to the user.
	DeleteGoal(ctx context.Context, userID, goalID string) error

	// GetMaxPriority returns the highest priority number currently used by this user's goals (or 0 if none).
	GetMaxPriority(ctx context.Context, userID string) (int, error)

	// CloseMonth executes the entire month-end rollover sweep in ONE atomic database transaction.
	// Go Concept - Closures passed as arguments:
	// The pure Allocate function is passed as a function argument into CloseMonth.
	// This cleanly separates database transaction orchestration from rollover allocation logic.
	CloseMonth(ctx context.Context, userID string, month time.Time, allocateFn func(leftover int64, goals []Goal) []Allocation) (CloseResult, error)

	// GetUnclosedPreviousMonths retrieves all (user_id, previous_month) pairs with budgets but no month_closings row.
	GetUnclosedPreviousMonths(ctx context.Context, previousMonth time.Time) ([]RolloverJob, error)
}

