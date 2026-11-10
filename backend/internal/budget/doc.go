// Package budget handles monthly category budget planning and tracking for Spendle.
//
// Future responsibilities:
// - Creating, updating, and querying monthly budgets per category (including custom categories)
// - All monetary figures stored and calculated in integer paise (int64)
// - Repository interface defining database queries for budgets using plain SQL
// - Service layer orchestrating budget thresholds and month-end unspent allocations
package budget

