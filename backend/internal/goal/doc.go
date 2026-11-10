// Package goal manages user savings goals and priority rankings for Spendle.
//
// Future responsibilities:
// - Creating, updating, and reordering savings goals with explicit priority ranks
// - Goal target amounts and current saved balances recorded strictly in integer paise (int64)
// - Month-end sweep service that takes unspent category budget and automatically allocates it into goals by priority
// - Handlers, services, and repository interfaces decoupling business rules from database access
package goal

