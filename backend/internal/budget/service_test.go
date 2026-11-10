package budget_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/budget"
)

// fakeBudgetRepository implements budget.Repository in-memory for unit testing.
type fakeBudgetRepository struct {
	categories []budget.Category
	budgets    []budget.BudgetProgress
}

func (f *fakeBudgetRepository) ListCategories(ctx context.Context, userID string) ([]budget.Category, error) {
	return f.categories, nil
}

func (f *fakeBudgetRepository) CreateCategory(ctx context.Context, userID, name string) (budget.Category, error) {
	return budget.Category{ID: "cat-1", UserID: userID, Name: name}, nil
}

func (f *fakeBudgetRepository) DeleteCategory(ctx context.Context, userID, categoryID string) error {
	return nil
}

func (f *fakeBudgetRepository) GetCategoryByID(ctx context.Context, userID, categoryID string) (budget.Category, error) {
	for _, c := range f.categories {
		if c.ID == categoryID && c.UserID == userID {
			return c, nil
		}
	}
	return budget.Category{}, budget.ErrCategoryNotFound
}

func (f *fakeBudgetRepository) UpsertBudget(ctx context.Context, userID, categoryID string, month time.Time, limitPaise int64) (budget.MonthlyBudget, error) {
	return budget.MonthlyBudget{
		ID:         "b-1",
		UserID:     userID,
		CategoryID: categoryID,
		Month:      month.Format("2006-01"),
		LimitPaise: limitPaise,
	}, nil
}

func (f *fakeBudgetRepository) GetBudgetProgress(ctx context.Context, userID string, monthStart, nextMonthStart time.Time) ([]budget.BudgetProgress, error) {
	return f.budgets, nil
}

// 1. Test: Month parsing: "2026-11" becomes the first day of November 2026, and a bad value returns an error.
func TestMonthParsing(t *testing.T) {
	t.Run("valid month 2026-11 becomes 2026-11-01 UTC", func(t *testing.T) {
		parsed, err := budget.ParseMonth("2026-11", false)
		if err != nil {
			t.Fatalf("unexpected error parsing valid month: %v", err)
		}

		expected := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
		if !parsed.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, parsed)
		}
	})

	t.Run("invalid month returns error", func(t *testing.T) {
		invalidInputs := []string{"invalid", "2026/11", "2026-13", "11-2026"}
		for _, input := range invalidInputs {
			_, err := budget.ParseMonth(input, false)
			if !errors.Is(err, budget.ErrInvalidMonth) {
				t.Errorf("expected ErrInvalidMonth for input %q, got: %v", input, err)
			}
		}
	})
}

// 2. Test: Budget left_paise = limit - spent, and overspending gives a negative number.
func TestBudgetCalculation_LeftAndOverspending(t *testing.T) {
	repo := &fakeBudgetRepository{
		budgets: []budget.BudgetProgress{
			{
				CategoryID:   "cat-food",
				CategoryName: "Food",
				LimitPaise:   100000, // ₹1,000.00
				SpentPaise:   150000, // ₹1,500.00 (overspent by ₹500.00)
				LeftPaise:    100000 - 150000, // -50000 paise
			},
			{
				CategoryID:   "cat-transport",
				CategoryName: "Transport",
				LimitPaise:   200000, // ₹2,000.00
				SpentPaise:   50000,  // ₹500.00
				LeftPaise:    200000 - 50000, // 150000 paise
			},
		},
	}

	svc := budget.NewService(repo)
	summary, err := svc.GetMonthlyBudgets(context.Background(), "user-1", "2026-11")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Food category overspent check
	food := summary.Budgets[0]
	if food.LeftPaise != -50000 {
		t.Errorf("expected left_paise to be -50000 for overspent category, got %d", food.LeftPaise)
	}

	// Overall month totals
	expectedTotalLimit := int64(300000)
	expectedTotalSpent := int64(200000)
	expectedTotalLeft := int64(100000)

	if summary.TotalLimitPaise != expectedTotalLimit {
		t.Errorf("expected total limit %d, got %d", expectedTotalLimit, summary.TotalLimitPaise)
	}
	if summary.TotalSpentPaise != expectedTotalSpent {
		t.Errorf("expected total spent %d, got %d", expectedTotalSpent, summary.TotalSpentPaise)
	}
	if summary.TotalLeftPaise != expectedTotalLeft {
		t.Errorf("expected total left %d, got %d", expectedTotalLeft, summary.TotalLeftPaise)
	}
}

