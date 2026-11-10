package goal_test

import (
	"testing"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/goal"
)

// Test 1: Allocate verifies leftover spills across goals by priority,
// stops when leftover runs out, and flags fully funded goals as completed.
func TestAllocate(t *testing.T) {
	goals := []goal.Goal{
		{
			ID:          "goal-1",
			Title:       "Emergency Fund",
			TargetPaise: 50000,
			SavedPaise:  20000, // needs 30000
			Priority:    1,
		},
		{
			ID:          "goal-2",
			Title:       "Bike",
			TargetPaise: 40000,
			SavedPaise:  0, // needs 40000
			Priority:    2,
		},
		{
			ID:          "goal-3",
			Title:       "MacBook",
			TargetPaise: 100000,
			SavedPaise:  0, // needs 100000
			Priority:    3,
		},
	}

	leftover := int64(50000)

	allocations := goal.Allocate(leftover, goals)

	// Goal 1 needs 30000 (has 20000, target 50000). It gets 30000 and is fully completed.
	// Remaining leftover is 20000.
	// Goal 2 needs 40000. It gets the remaining 20000 and is NOT completed.
	// Remaining leftover is 0.
	// Goal 3 receives nothing and is not in allocations.
	if len(allocations) != 2 {
		t.Fatalf("expected 2 allocations, got %d", len(allocations))
	}

	// Verify Goal 1
	if allocations[0].GoalID != "goal-1" {
		t.Errorf("expected first allocation for goal-1, got %s", allocations[0].GoalID)
	}
	if allocations[0].AmountPaise != 30000 {
		t.Errorf("expected goal-1 to receive 30000 paise, got %d", allocations[0].AmountPaise)
	}
	if !allocations[0].GoalCompleted {
		t.Errorf("expected goal-1 to be completed, got false")
	}

	// Verify Goal 2
	if allocations[1].GoalID != "goal-2" {
		t.Errorf("expected second allocation for goal-2, got %s", allocations[1].GoalID)
	}
	if allocations[1].AmountPaise != 20000 {
		t.Errorf("expected goal-2 to receive 20000 paise, got %d", allocations[1].AmountPaise)
	}
	if allocations[1].GoalCompleted {
		t.Errorf("expected goal-2 to not be completed, got true")
	}
}

