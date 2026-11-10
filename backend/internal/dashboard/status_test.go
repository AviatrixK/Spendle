package dashboard_test

import (
	"testing"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/dashboard"
)

// Test 4: The dashboard status function for profit, loss, and break_even.
func TestCalculateStatus(t *testing.T) {
	tests := []struct {
		name     string
		netPaise int64
		expected string
	}{
		{
			name:     "positive net paise indicates profit",
			netPaise: 75000,
			expected: "profit",
		},
		{
			name:     "negative net paise indicates loss",
			netPaise: -25000,
			expected: "loss",
		},
		{
			name:     "zero net paise indicates break_even",
			netPaise: 0,
			expected: "break_even",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dashboard.CalculateStatus(tt.netPaise)
			if got != tt.expected {
				t.Errorf("expected status %q, got %q", tt.expected, got)
			}
		})
	}
}

