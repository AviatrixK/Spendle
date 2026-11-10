package goal

// Allocate assigns leftover unspent budget across active goals by priority.
// Go Concept - Pure Functions & Closures:
// Allocate is a pure deterministic algorithm with zero side effects. It can be passed as a
// closure argument into the repository transaction method, separating domain logic from database I/O.
//
// Rules:
// 1. Iterates active goals ordered by priority.
// 2. Gives each goal min(goal remaining, leftover left).
// 3. Flags goal as completed if its new total saved reaches or exceeds target.
// 4. Returns only allocations with AmountPaise > 0.
func Allocate(leftover int64, goals []Goal) []Allocation {
	if leftover <= 0 {
		return nil
	}

	remainingLeftover := leftover
	var allocations []Allocation

	for _, g := range goals {
		if remainingLeftover <= 0 {
			break
		}

		needed := g.TargetPaise - g.SavedPaise
		if needed <= 0 {
			continue // Already reached target
		}

		give := needed
		if give > remainingLeftover {
			give = remainingLeftover
		}

		remainingLeftover -= give
		completed := (g.SavedPaise + give) >= g.TargetPaise

		allocations = append(allocations, Allocation{
			GoalID:        g.ID,
			AmountPaise:   give,
			GoalCompleted: completed,
		})
	}

	return allocations
}

