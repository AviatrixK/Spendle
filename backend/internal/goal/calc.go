package goal

import (
	"time"
)

// RemainingPaise calculates remaining funds needed: max(target - saved, 0).
// Pure function: operates strictly on inputs without external database or state.
func RemainingPaise(target, saved int64) int64 {
	rem := target - saved
	if rem < 0 {
		return 0
	}
	return rem
}

// ProgressPercent calculates goal progress as an integer 0-100, strictly capped at 100.
// Pure function: deterministic and easy to unit test.
func ProgressPercent(target, saved int64) int {
	if target <= 0 || saved <= 0 {
		return 0
	}
	pct := (saved * 100) / target
	if pct > 100 {
		return 100
	}
	return int(pct)
}

// MonthlyNeededPaise calculates required monthly savings given a target deadline.
// Formula: ceil(remaining / months_left) where months_left >= 1 calendar month.
// Returns nil if no deadline is specified.
func MonthlyNeededPaise(remaining int64, deadline *time.Time, now time.Time) *int64 {
	if deadline == nil {
		return nil
	}
	if remaining <= 0 {
		zero := int64(0)
		return &zero
	}

	// Compute calendar months from current month to deadline month
	monthsLeft := (deadline.Year()-now.Year())*12 + int(deadline.Month()-now.Month())
	if monthsLeft < 1 {
		monthsLeft = 1
	}

	// Integer division ceiling: ceil(A / B) = (A + B - 1) / B
	needed := (remaining + int64(monthsLeft) - 1) / int64(monthsLeft)
	return &needed
}

// ETAMonth calculates the projected completion month in "YYYY-MM" format.
// Formula: current month + ceil(remaining / monthly_amount) months.
// Returns nil if monthly amount is missing, zero, or remaining is 0.
func ETAMonth(remaining int64, monthlyAmount *int64, now time.Time) *string {
	if monthlyAmount == nil || *monthlyAmount <= 0 || remaining <= 0 {
		return nil
	}

	// Integer division ceiling
	monthsNeeded := (remaining + *monthlyAmount - 1) / *monthlyAmount
	targetMonth := now.AddDate(0, int(monthsNeeded), 0)
	formatted := targetMonth.Format("2006-01")
	return &formatted
}

