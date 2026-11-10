package transaction

import "errors"

// Sentinel errors for transaction management.
var (
	// ErrInvalidType is returned when the transaction type is neither 'income' nor 'expense'.
	ErrInvalidType = errors.New("transaction type must be 'income' or 'expense'")

	// ErrInvalidAmount is returned when amount_paise is zero or negative.
	ErrInvalidAmount = errors.New("amount must be greater than zero paise")

	// ErrInvalidPaymentMode is returned when payment_mode is not one of upi, cash, card, netbanking, other.
	ErrInvalidPaymentMode = errors.New("invalid payment mode; allowed values are upi, cash, card, netbanking, other")

	// ErrExpenseCategoryRequired is returned when an expense transaction is submitted without a category_id.
	ErrExpenseCategoryRequired = errors.New("category is required for expense transactions")

	// ErrCategoryNotFound is returned when the specified category_id does not exist or belong to the user.
	ErrCategoryNotFound = errors.New("category not found")

	// ErrTransactionNotFound is returned when the requested transaction does not exist or belong to the user.
	ErrTransactionNotFound = errors.New("transaction not found")

	// ErrNoteTooLong is returned when note exceeds 200 characters.
	ErrNoteTooLong = errors.New("note cannot exceed 200 characters")

	// ErrInvalidOccurredOn is returned when occurred_on is not a valid YYYY-MM-DD date.
	ErrInvalidOccurredOn = errors.New("invalid occurred_on date format; expected YYYY-MM-DD")

	// ErrInvalidMonth is returned when month filter is not in YYYY-MM format.
	ErrInvalidMonth = errors.New("invalid month format; expected YYYY-MM")
)

