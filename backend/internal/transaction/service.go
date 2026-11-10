package transaction

import (
	"context"
	"strings"
	"time"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/budget"
)

// CategoryChecker is an interface for verifying category ownership across packages.
type CategoryChecker interface {
	GetCategoryByID(ctx context.Context, userID, categoryID string) (budget.Category, error)
}

// Service contains business logic operations for financial transactions.
type Service struct {
	repo            Repository
	categoryChecker CategoryChecker
}

// NewService constructs a transaction Service instance with its dependencies.
func NewService(repo Repository, categoryChecker CategoryChecker) *Service {
	return &Service{
		repo:            repo,
		categoryChecker: categoryChecker,
	}
}

// CreateInput defines parameters required to create an income or expense transaction.
type CreateInput struct {
	Type        string
	AmountPaise int64
	CategoryID  *string
	PaymentMode string
	Note        string
	OccurredOn  string
}

// Create validates and saves a new financial transaction.
func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (Transaction, error) {
	txType := strings.ToLower(strings.TrimSpace(input.Type))
	if txType != "income" && txType != "expense" {
		return Transaction{}, ErrInvalidType
	}

	// Strict Monetary Rule: transaction amount must be positive integer paise
	if input.AmountPaise <= 0 {
		return Transaction{}, ErrInvalidAmount
	}

	// Default payment mode to "upi" if not specified
	paymentMode := strings.ToLower(strings.TrimSpace(input.PaymentMode))
	if paymentMode == "" {
		paymentMode = "upi"
	}
	switch paymentMode {
	case "upi", "cash", "card", "netbanking", "other":
		// valid payment mode
	default:
		return Transaction{}, ErrInvalidPaymentMode
	}

	note := strings.TrimSpace(input.Note)
	if len(note) > 200 {
		return Transaction{}, ErrNoteTooLong
	}

	// OccurredOn defaults to today if omitted
	occurredOn := strings.TrimSpace(input.OccurredOn)
	if occurredOn == "" {
		occurredOn = time.Now().UTC().Format("2006-01-02")
	} else {
		if _, err := time.Parse("2006-01-02", occurredOn); err != nil {
			return Transaction{}, ErrInvalidOccurredOn
		}
	}

	// Category rules:
	// - An EXPENSE requires a category belonging to the user (400 if missing, 404 if not found).
	// - An INCOME may omit the category.
	var catID *string
	var catName *string
	if input.CategoryID != nil && strings.TrimSpace(*input.CategoryID) != "" {
		trimmedID := strings.TrimSpace(*input.CategoryID)
		category, err := s.categoryChecker.GetCategoryByID(ctx, userID, trimmedID)
		if err != nil {
			return Transaction{}, ErrCategoryNotFound
		}
		catID = &category.ID
		catName = &category.Name
	} else if txType == "expense" {
		return Transaction{}, ErrExpenseCategoryRequired
	}

	tx := Transaction{
		UserID:       userID,
		Type:         txType,
		AmountPaise:  input.AmountPaise,
		CategoryID:   catID,
		CategoryName: catName,
		PaymentMode:  paymentMode,
		Note:         note,
		OccurredOn:   occurredOn,
	}

	created, err := s.repo.CreateTransaction(ctx, tx)
	if err != nil {
		return Transaction{}, err
	}
	created.CategoryName = catName
	return created, nil
}

// List retrieves transactions for a user within a calendar month with pagination limit.
func (s *Service) List(ctx context.Context, userID, monthStr string, limit int) ([]Transaction, error) {
	if limit <= 0 {
		limit = 50
	} else if limit > 200 {
		limit = 200
	}

	monthStart, err := parseMonth(monthStr)
	if err != nil {
		return nil, err
	}
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	return s.repo.ListTransactions(ctx, userID, monthStart, nextMonthStart, limit)
}

// Delete removes a transaction belonging to the authenticated user.
func (s *Service) Delete(ctx context.Context, userID, transactionID string) error {
	return s.repo.DeleteTransaction(ctx, userID, transactionID)
}

// GetSummary calculates total income, expenses, and net profit/loss for a calendar month.
func (s *Service) GetSummary(ctx context.Context, userID, monthStr string) (Summary, error) {
	monthStart, err := parseMonth(monthStr)
	if err != nil {
		return Summary{}, err
	}
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	return s.repo.GetMonthlySummary(ctx, userID, monthStart, nextMonthStart)
}

// parseMonth parses "YYYY-MM" into a UTC time.Time at the 1st of the month, defaulting to current month if empty.
func parseMonth(monthStr string) (time.Time, error) {
	monthStr = strings.TrimSpace(monthStr)
	if monthStr == "" {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}

	t, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return time.Time{}, ErrInvalidMonth
	}

	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

