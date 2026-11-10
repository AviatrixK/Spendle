package dashboard

import (
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/goal"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/transaction"
)

// BudgetItem represents spent vs left progress for a single category on the dashboard.
type BudgetItem struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	LimitPaise   int64  `json:"limit_paise"`
	SpentPaise   int64  `json:"spent_paise"`
	LeftPaise    int64  `json:"left_paise"`
}

// BudgetsSummary aggregates monthly budget limits and category breakdowns for the dashboard.
type BudgetsSummary struct {
	TotalLimitPaise int64        `json:"total_limit_paise"`
	TotalSpentPaise int64        `json:"total_spent_paise"`
	TotalLeftPaise  int64        `json:"total_left_paise"`
	Items           []BudgetItem `json:"items"`
}

// DashboardResponse encapsulates the full monthly dashboard payload.
type DashboardResponse struct {
	Month   string              `json:"month"`
	Status  string              `json:"status"` // "profit", "loss", or "break_even"
	Summary transaction.Summary `json:"summary"`
	Budgets BudgetsSummary      `json:"budgets"`
	Goals   []goal.GoalResponse `json:"goals"`
}

