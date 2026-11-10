-- Migration: 000001_init.up.sql
-- Description: Core Spendle schema for users, categories, budgets, transactions, goals, and month closings.
-- Strict Monetary Rule: All amounts are stored strictly as integer paise (BIGINT). Zero floats.

-- 1. Users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_users_email_lowercase CHECK (email = lower(email))
);

-- 2. Categories
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_categories_user_name UNIQUE (user_id, name)
);

-- 3. Monthly Budgets
-- Month is strictly the first day of the month (enforced via CHECK constraint)
CREATE TABLE IF NOT EXISTS monthly_budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    month DATE NOT NULL,
    limit_paise BIGINT NOT NULL CHECK (limit_paise >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_monthly_budgets_first_of_month CHECK (EXTRACT(DAY FROM month) = 1),
    CONSTRAINT uq_monthly_budgets_user_category_month UNIQUE (user_id, category_id, month)
);

-- 4. Transactions
-- One-tap logging for income and expenses across UPI, Cash, Card, Netbanking, Other
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
    amount_paise BIGINT NOT NULL CHECK (amount_paise > 0),
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    payment_mode VARCHAR(20) NOT NULL CHECK (payment_mode IN ('upi', 'cash', 'card', 'netbanking', 'other')),
    note TEXT,
    occurred_on DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 5. Goals
-- Ranked savings targets (priority 1 = highest)
CREATE TABLE IF NOT EXISTS goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    target_paise BIGINT NOT NULL CHECK (target_paise > 0),
    saved_paise BIGINT NOT NULL DEFAULT 0 CHECK (saved_paise >= 0),
    deadline DATE,
    monthly_amount_paise BIGINT CHECK (monthly_amount_paise IS NULL OR monthly_amount_paise > 0),
    priority INT NOT NULL CHECK (priority >= 1),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused', 'completed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 6. Goal Contributions
-- Contributions via manual deposit or month-end rollover sweep
CREATE TABLE IF NOT EXISTS goal_contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount_paise BIGINT NOT NULL CHECK (amount_paise > 0),
    source VARCHAR(20) NOT NULL CHECK (source IN ('manual', 'rollover')),
    month DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 7. Month Closings
-- Makes month-end rollover sweeps idempotent per user per month
CREATE TABLE IF NOT EXISTS month_closings (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    month DATE NOT NULL CHECK (EXTRACT(DAY FROM month) = 1),
    closed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, month)
);

-- Common Query Indexes
CREATE INDEX IF NOT EXISTS idx_transactions_user_occurred_on ON transactions (user_id, occurred_on);
CREATE INDEX IF NOT EXISTS idx_transactions_user_category_occurred_on ON transactions (user_id, category_id, occurred_on);
CREATE INDEX IF NOT EXISTS idx_monthly_budgets_user_month ON monthly_budgets (user_id, month);
CREATE INDEX IF NOT EXISTS idx_goals_user_priority ON goals (user_id, priority);

