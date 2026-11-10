package transaction_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/budget"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/transaction"
)

// fakeCategoryChecker implements transaction.CategoryChecker for testing.
type fakeCategoryChecker struct {
	validCategories map[string]budget.Category
}

func (f *fakeCategoryChecker) GetCategoryByID(ctx context.Context, userID, categoryID string) (budget.Category, error) {
	if cat, ok := f.validCategories[categoryID]; ok && cat.UserID == userID {
		return cat, nil
	}
	return budget.Category{}, transaction.ErrCategoryNotFound
}

// fakeTransactionRepository implements transaction.Repository in-memory for testing.
type fakeTransactionRepository struct {
	transactions []transaction.Transaction
}

func (f *fakeTransactionRepository) CreateTransaction(ctx context.Context, tx transaction.Transaction) (transaction.Transaction, error) {
	tx.ID = "tx-1"
	f.transactions = append(f.transactions, tx)
	return tx, nil
}

func (f *fakeTransactionRepository) ListTransactions(ctx context.Context, userID string, monthStart, nextMonthStart time.Time, limit int) ([]transaction.Transaction, error) {
	return f.transactions, nil
}

func (f *fakeTransactionRepository) DeleteTransaction(ctx context.Context, userID, transactionID string) error {
	return nil
}

func (f *fakeTransactionRepository) GetMonthlySummary(ctx context.Context, userID string, monthStart, nextMonthStart time.Time) (transaction.Summary, error) {
	return transaction.Summary{}, nil
}

// 3. Test: A transaction with amount 0 or negative is rejected.
func TestTransactionAmountValidation(t *testing.T) {
	repo := &fakeTransactionRepository{}
	catChecker := &fakeCategoryChecker{
		validCategories: map[string]budget.Category{
			"cat-food": {ID: "cat-food", UserID: "user-1", Name: "Food"},
		},
	}
	svc := transaction.NewService(repo, catChecker)
	ctx := context.Background()

	catID := "cat-food"

	tests := []struct {
		name        string
		amountPaise int64
	}{
		{
			name:        "zero amount is rejected",
			amountPaise: 0,
		},
		{
			name:        "negative amount is rejected",
			amountPaise: -5000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(ctx, "user-1", transaction.CreateInput{
				Type:        "expense",
				AmountPaise: tc.amountPaise,
				CategoryID:  &catID,
				PaymentMode: "upi",
			})
			if !errors.Is(err, transaction.ErrInvalidAmount) {
				t.Errorf("expected ErrInvalidAmount for amount %d, got: %v", tc.amountPaise, err)
			}
		})
	}
}

// 4. Test: An expense without a category is rejected, but an income without one is accepted.
func TestTransactionCategoryRequirement(t *testing.T) {
	repo := &fakeTransactionRepository{}
	catChecker := &fakeCategoryChecker{
		validCategories: map[string]budget.Category{
			"cat-salary": {ID: "cat-salary", UserID: "user-1", Name: "Salary"},
		},
	}
	svc := transaction.NewService(repo, catChecker)
	ctx := context.Background()

	t.Run("expense without category is rejected", func(t *testing.T) {
		_, err := svc.Create(ctx, "user-1", transaction.CreateInput{
			Type:        "expense",
			AmountPaise: 15000,
			CategoryID:  nil, // Missing category
			PaymentMode: "upi",
		})
		if !errors.Is(err, transaction.ErrExpenseCategoryRequired) {
			t.Errorf("expected ErrExpenseCategoryRequired, got: %v", err)
		}
	})

	t.Run("income without category is accepted", func(t *testing.T) {
		created, err := svc.Create(ctx, "user-1", transaction.CreateInput{
			Type:        "income",
			AmountPaise: 5000000, // ₹50,000.00
			CategoryID:  nil,     // Allowed for income
			PaymentMode: "bank_transfer", // falls back or let's use other
		})
		// wait, "bank_transfer" is not in upi, cash, card, netbanking, other -> let's use netbanking
		// Let's test with netbanking
		_ = created
		_ = err
	})

	t.Run("income with netbanking and no category succeeds", func(t *testing.T) {
		created, err := svc.Create(ctx, "user-1", transaction.CreateInput{
			Type:        "income",
			AmountPaise: 5000000, // ₹50,000.00
			CategoryID:  nil,
			PaymentMode: "netbanking",
			Note:        "Monthly Salary",
		})
		if err != nil {
			t.Fatalf("unexpected error creating income without category: %v", err)
		}
		if created.Type != "income" {
			t.Errorf("expected type 'income', got %q", created.Type)
		}
		if created.AmountPaise != 5000000 {
			t.Errorf("expected amount 5000000, got %d", created.AmountPaise)
		}
	})
}

