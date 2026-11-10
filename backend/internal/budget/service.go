package budget

import (
	"context"
	"strings"
	"time"
)

// Service coordinates business logic for user categories and monthly budget limits.
type Service struct {
	repo Repository
}

// NewService constructs a budget Service instance.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListCategories returns all categories belonging to the user as [{id, name}].
func (s *Service) ListCategories(ctx context.Context, userID string) ([]CategoryResponse, error) {
	categories, err := s.repo.ListCategories(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]CategoryResponse, 0, len(categories))
	for _, c := range categories {
		responses = append(responses, CategoryResponse{
			ID:   c.ID,
			Name: c.Name,
		})
	}
	return responses, nil
}

// CreateCategory validates that the category name is trimmed and 1-40 characters.
func (s *Service) CreateCategory(ctx context.Context, userID, name string) (CategoryResponse, error) {
	name = strings.TrimSpace(name)
	if len(name) < 1 || len(name) > 40 {
		return CategoryResponse{}, ErrInvalidCategoryName
	}

	category, err := s.repo.CreateCategory(ctx, userID, name)
	if err != nil {
		return CategoryResponse{}, err
	}

	return CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
	}, nil
}

// DeleteCategory deletes a category ensuring it belongs to the user.
func (s *Service) DeleteCategory(ctx context.Context, userID, categoryID string) error {
	return s.repo.DeleteCategory(ctx, userID, categoryID)
}

// SetBudget validates the limit, verifies category ownership, and upserts the budget row.
func (s *Service) SetBudget(ctx context.Context, userID, categoryID, monthStr string, limitPaise int64) (MonthlyBudget, error) {
	if limitPaise < 0 {
		return MonthlyBudget{}, ErrInvalidLimit
	}

	month, err := ParseMonth(monthStr, false)
	if err != nil {
		return MonthlyBudget{}, err
	}

	// Verify the category belongs to this user (404 ErrCategoryNotFound if missing)
	if _, err := s.repo.GetCategoryByID(ctx, userID, categoryID); err != nil {
		return MonthlyBudget{}, err
	}

	return s.repo.UpsertBudget(ctx, userID, categoryID, month, limitPaise)
}

// GetMonthlyBudgets calculates per-category spent/left progress and monthly totals.
func (s *Service) GetMonthlyBudgets(ctx context.Context, userID, monthStr string) (MonthBudgetSummary, error) {
	monthStart, err := ParseMonth(monthStr, true)
	if err != nil {
		return MonthBudgetSummary{}, err
	}
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	budgets, err := s.repo.GetBudgetProgress(ctx, userID, monthStart, nextMonthStart)
	if err != nil {
		return MonthBudgetSummary{}, err
	}

	var totalLimit, totalSpent int64
	for _, b := range budgets {
		totalLimit += b.LimitPaise
		totalSpent += b.SpentPaise
	}

	// Guarantee empty slice instead of null
	if budgets == nil {
		budgets = make([]BudgetProgress, 0)
	}

	return MonthBudgetSummary{
		Month:           monthStart.Format("2006-01"),
		TotalLimitPaise: totalLimit,
		TotalSpentPaise: totalSpent,
		TotalLeftPaise:  totalLimit - totalSpent,
		Budgets:         budgets,
	}, nil
}

// ParseMonth converts a "YYYY-MM" string to the first day of that month in UTC.
// If allowDefault is true and monthStr is empty, it defaults to the current calendar month.
func ParseMonth(monthStr string, allowDefault bool) (time.Time, error) {
	monthStr = strings.TrimSpace(monthStr)
	if monthStr == "" {
		if allowDefault {
			now := time.Now().UTC()
			return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), nil
		}
		return time.Time{}, ErrInvalidMonth
	}

	t, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return time.Time{}, ErrInvalidMonth
	}

	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}
