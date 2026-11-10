package dashboard

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/budget"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/goal"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/transaction"
)

// Go Concept - Consumer-Defined Interfaces:
// In Go, interfaces should be defined where they are used (the consumer), not where they are implemented.
// This decouples the dashboard package from concrete services in budget, transaction, and goal.
// Dashboard defines only the minimal contracts it requires, making testing with mock implementations trivial.

// BudgetsReader defines the consumer contract for retrieving monthly budget progress.
type BudgetsReader interface {
	GetMonthlyBudgets(ctx context.Context, userID, monthStr string) (budget.MonthBudgetSummary, error)
}

// SummaryReader defines the consumer contract for retrieving monthly financial summaries.
type SummaryReader interface {
	GetSummary(ctx context.Context, userID, monthStr string) (transaction.Summary, error)
}

// GoalsReader defines the consumer contract for listing user savings goals.
type GoalsReader interface {
	ListGoals(ctx context.Context, userID string) ([]goal.GoalResponse, error)
}

// BudgetServiceAdapter adapts *budget.Service to the BudgetsReader interface.
type BudgetServiceAdapter struct {
	svc *budget.Service
}

func NewBudgetServiceAdapter(svc *budget.Service) *BudgetServiceAdapter {
	return &BudgetServiceAdapter{svc: svc}
}

func (a *BudgetServiceAdapter) GetMonthlyBudgets(ctx context.Context, userID, monthStr string) (budget.MonthBudgetSummary, error) {
	return a.svc.GetMonthlyBudgets(ctx, userID, monthStr)
}

// TransactionServiceAdapter adapts *transaction.Service to the SummaryReader interface.
type TransactionServiceAdapter struct {
	svc *transaction.Service
}

func NewTransactionServiceAdapter(svc *transaction.Service) *TransactionServiceAdapter {
	return &TransactionServiceAdapter{svc: svc}
}

func (a *TransactionServiceAdapter) GetSummary(ctx context.Context, userID, monthStr string) (transaction.Summary, error) {
	return a.svc.GetSummary(ctx, userID, monthStr)
}

// GoalServiceAdapter adapts *goal.Service to the GoalsReader interface.
type GoalServiceAdapter struct {
	svc *goal.Service
}

func NewGoalServiceAdapter(svc *goal.Service) *GoalServiceAdapter {
	return &GoalServiceAdapter{svc: svc}
}

func (a *GoalServiceAdapter) ListGoals(ctx context.Context, userID string) ([]goal.GoalResponse, error) {
	return a.svc.ListGoals(ctx, userID)
}

// Service orchestrates loading dashboard data across budgets, transactions, and goals concurrently.
type Service struct {
	budgets BudgetsReader
	summary SummaryReader
	goals   GoalsReader
}

// NewService constructs a dashboard Service with consumer interface dependencies.
func NewService(budgets BudgetsReader, summary SummaryReader, goals GoalsReader) *Service {
	return &Service{
		budgets: budgets,
		summary: summary,
		goals:   goals,
	}
}

// GetDashboard loads monthly metrics concurrently using errgroup and a 3-second context timeout.
func (s *Service) GetDashboard(ctx context.Context, userID, monthStr string) (DashboardResponse, error) {
	// Go Concept - Context Timeout:
	// context.WithTimeout creates a deadline. If concurrent calls take longer than 3 seconds,
	// the timeoutCtx is cancelled automatically, freeing database connections and goroutines.
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Go Concept - errgroup:
	// errgroup.WithContext coordinates concurrent goroutines that return an error.
	// If any subtask returns an error, errgroup cancels the derived context (gCtx), aborting
	// sibling goroutines, and g.Wait() returns the first error encountered.
	g, gCtx := errgroup.WithContext(timeoutCtx)

	var (
		txSummary     transaction.Summary
		budgetSummary budget.MonthBudgetSummary
		goalsList     []goal.GoalResponse
	)

	// Subtask 1: Fetch transaction summary concurrently
	g.Go(func() error {
		var err error
		txSummary, err = s.summary.GetSummary(gCtx, userID, monthStr)
		return err
	})

	// Subtask 2: Fetch budget progress concurrently
	g.Go(func() error {
		var err error
		budgetSummary, err = s.budgets.GetMonthlyBudgets(gCtx, userID, monthStr)
		return err
	})

	// Subtask 3: Fetch goals list concurrently
	g.Go(func() error {
		var err error
		goalsList, err = s.goals.ListGoals(gCtx, userID)
		return err
	})

	// Wait for all 3 concurrent goroutines to complete
	if err := g.Wait(); err != nil {
		return DashboardResponse{}, err
	}

	// Map budget progress items into dashboard BudgetItem structure
	items := make([]BudgetItem, 0, len(budgetSummary.Budgets))
	for _, b := range budgetSummary.Budgets {
		items = append(items, BudgetItem{
			CategoryID:   b.CategoryID,
			CategoryName: b.CategoryName,
			LimitPaise:   b.LimitPaise,
			SpentPaise:   b.SpentPaise,
			LeftPaise:    b.LeftPaise,
		})
	}

	if goalsList == nil {
		goalsList = make([]goal.GoalResponse, 0)
	}

	return DashboardResponse{
		Month:   monthStr,
		Status:  CalculateStatus(txSummary.NetPaise),
		Summary: txSummary,
		Budgets: BudgetsSummary{
			TotalLimitPaise: budgetSummary.TotalLimitPaise,
			TotalSpentPaise: budgetSummary.TotalSpentPaise,
			TotalLeftPaise:  budgetSummary.TotalLeftPaise,
			Items:           items,
		},
		Goals: goalsList,
	}, nil
}

