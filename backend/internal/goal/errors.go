package goal

import "errors"

var (
	ErrInvalidTitle             = errors.New("title must be between 1 and 60 characters")
	ErrInvalidTarget            = errors.New("target_paise must be greater than 0")
	ErrConflictingTimeline      = errors.New("cannot specify both deadline and monthly_amount_paise")
	ErrInvalidDeadline          = errors.New("deadline must be in YYYY-MM-DD format")
	ErrInvalidMonthlyAmount     = errors.New("monthly_amount_paise must be greater than 0")
	ErrInvalidPriority          = errors.New("priority must be at least 1")
	ErrInvalidStatus            = errors.New("status must be 'active', 'paused', or 'completed'")
	ErrGoalNotFound             = errors.New("goal not found")
	ErrFutureMonth              = errors.New("cannot close future month")
	ErrInvalidMonth             = errors.New("month must be formatted as YYYY-MM")
)

