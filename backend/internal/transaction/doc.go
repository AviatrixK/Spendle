// Package transaction handles logging and querying of income and expense transactions.
//
// Future responsibilities:
// - One-tap creation of expense and income transactions
// - Recording payment modes: UPI, Cash, Card
// - Managing categories and category relationships
// - Strict handling of transaction amounts as integer paise (int64 / BIGINT)
// - Plain SQL repository implementations with repository interfaces for unit testing
package transaction

