package goal

import "time"

// Goal represents a user's savings target with a priority ranking.
// Strict Monetary Precision: Target and saved amounts are stored strictly in integer paise (int64).
type Goal struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	Title              string    `json:"title"`
	TargetPaise        int64     `json:"target_paise"`
	SavedPaise         int64     `json:"saved_paise"`
	Deadline           *string   `json:"deadline"`             // Format: "YYYY-MM-DD", optional
	MonthlyAmountPaise *int64    `json:"monthly_amount_paise"` // optional int64 paise
	Priority           int       `json:"priority"`
	Status             string    `json:"status"` // "active", "paused", "completed"
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// GoalResponse represents the public JSON structure for a goal, including all computed fields.
type GoalResponse struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	Title              string    `json:"title"`
	TargetPaise        int64     `json:"target_paise"`
	SavedPaise         int64     `json:"saved_paise"`
	Deadline           *string   `json:"deadline"`
	MonthlyAmountPaise *int64    `json:"monthly_amount_paise"`
	Priority           int       `json:"priority"`
	Status             string    `json:"status"`
	RemainingPaise     int64     `json:"remaining_paise"`
	ProgressPercent    int       `json:"progress_percent"`
	MonthlyNeededPaise *int64    `json:"monthly_needed_paise"`
	EtaMonth           *string   `json:"eta_month"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Allocation represents funds assigned to a specific goal during a month-end rollover sweep.
type Allocation struct {
	GoalID        string `json:"goal_id"`
	AmountPaise   int64  `json:"amount_paise"`
	GoalCompleted bool   `json:"goal_completed"`
}

// CloseResult represents the outcome of a month-end rollover sweep.
type CloseResult struct {
	AlreadyClosed bool  `json:"already_closed"`
	MovedPaise    int64 `json:"moved_paise"`
	GoalsFunded   int   `json:"goals_funded"`
}

// RolloverJob represents an unclosed month job queued for a worker goroutine.
type RolloverJob struct {
	UserID string    `json:"user_id"`
	Month  time.Time `json:"month"`
}

