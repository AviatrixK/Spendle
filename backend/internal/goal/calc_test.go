package goal_test

import (
	"testing"
	"time"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/goal"
)

// Test 2: monthly_needed_paise with a deadline, and eta_month with a monthly amount
// (including the "at least 1 month" rule).
func TestMonthlyNeededAndETAMonth(t *testing.T) {
	// Base date: 2026-11-10
	now := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)

	t.Run("monthly_needed_paise with future deadline", func(t *testing.T) {
		// Deadline 5 months away: 2027-04-01 -> calendar months = 5
		deadline := time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC)
		remaining := int64(50000)

		needed := goal.MonthlyNeededPaise(remaining, &deadline, now)
		if needed == nil {
			t.Fatal("expected monthly_needed_paise to not be nil")
		}
		// ceil(50000 / 5) = 10000
		if *needed != 10000 {
			t.Errorf("expected 10000 paise, got %d", *needed)
		}
	})

	t.Run("monthly_needed_paise with current month deadline enforces at least 1 month", func(t *testing.T) {
		// Deadline in the current month: 2026-11-25 -> calendar diff is 0, clamped to at least 1
		deadline := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
		remaining := int64(15000)

		needed := goal.MonthlyNeededPaise(remaining, &deadline, now)
		if needed == nil {
			t.Fatal("expected monthly_needed_paise to not be nil")
		}
		// ceil(15000 / 1) = 15000
		if *needed != 15000 {
			t.Errorf("expected 15000 paise, got %d", *needed)
		}
	})

	t.Run("monthly_needed_paise with past deadline clamps to at least 1 month", func(t *testing.T) {
		// Deadline in a past month: 2026-09-01 -> calendar diff is -2, clamped to at least 1
		deadline := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		remaining := int64(20000)

		needed := goal.MonthlyNeededPaise(remaining, &deadline, now)
		if needed == nil {
			t.Fatal("expected monthly_needed_paise to not be nil")
		}
		// ceil(20000 / 1) = 20000
		if *needed != 20000 {
			t.Errorf("expected 20000 paise, got %d", *needed)
		}
	})

	t.Run("eta_month with monthly amount", func(t *testing.T) {
		// Remaining 90000 paise, monthly amount 40000 paise -> ceil(90000 / 40000) = 3 months
		// Current month: 2026-11 + 3 months = 2027-02
		remaining := int64(90000)
		monthlyAmount := int64(40000)

		eta := goal.ETAMonth(remaining, &monthlyAmount, now)
		if eta == nil {
			t.Fatal("expected eta_month to not be nil")
		}
		if *eta != "2027-02" {
			t.Errorf("expected '2027-02', got %q", *eta)
		}
	})

	t.Run("eta_month when remaining is 0 returns nil", func(t *testing.T) {
		remaining := int64(0)
		monthlyAmount := int64(25000)

		eta := goal.ETAMonth(remaining, &monthlyAmount, now)
		if eta != nil {
			t.Errorf("expected nil eta when remaining is 0, got %q", *eta)
		}
	})
}

// Test 3: progress_percent is capped at 100.
func TestProgressPercent(t *testing.T) {
	tests := []struct {
		name     string
		target   int64
		saved    int64
		expected int
	}{
		{name: "0% progress", target: 10000, saved: 0, expected: 0},
		{name: "25% progress", target: 10000, saved: 2500, expected: 25},
		{name: "50% progress", target: 10000, saved: 5000, expected: 50},
		{name: "100% progress", target: 10000, saved: 10000, expected: 100},
		{name: "Capped at 100% when saved exceeds target", target: 10000, saved: 15000, expected: 100},
		{name: "Negative saved returns 0", target: 10000, saved: -500, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := goal.ProgressPercent(tt.target, tt.saved)
			if got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

