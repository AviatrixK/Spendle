package budget

import "errors"

// Sentinel errors for categories and monthly budgets.
var (
	// ErrCategoryExists is returned when attempting to create a duplicate category name for the same user.
	ErrCategoryExists = errors.New("category with this name already exists")

	// ErrCategoryNotFound is returned when a category does not exist or does not belong to the user.
	ErrCategoryNotFound = errors.New("category not found")

	// ErrInvalidMonth is returned when month string is not in YYYY-MM format.
	ErrInvalidMonth = errors.New("invalid month format; expected YYYY-MM")

	// ErrInvalidLimit is returned when attempting to set a negative budget limit.
	ErrInvalidLimit = errors.New("budget limit must be greater than or equal to zero paise")

	// ErrInvalidCategoryName is returned when category name is not 1-40 characters.
	ErrInvalidCategoryName = errors.New("category name must be between 1 and 40 characters")
)
