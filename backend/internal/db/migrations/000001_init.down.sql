-- Migration: 000001_init.down.sql
-- Description: Cleanly drops all tables and indexes created in 000001_init.up.sql

DROP INDEX IF EXISTS idx_goals_user_priority;
DROP INDEX IF EXISTS idx_monthly_budgets_user_month;
DROP INDEX IF EXISTS idx_transactions_user_category_occurred_on;
DROP INDEX IF EXISTS idx_transactions_user_occurred_on;

DROP TABLE IF EXISTS month_closings CASCADE;
DROP TABLE IF EXISTS goal_contributions CASCADE;
DROP TABLE IF EXISTS goals CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS monthly_budgets CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS users CASCADE;

