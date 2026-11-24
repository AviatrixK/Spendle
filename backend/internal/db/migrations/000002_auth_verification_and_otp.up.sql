-- Migration: 000002_auth_verification_and_otp.up.sql
-- Description: Adds email verification status, session invalidation token version, Google OAuth identity, and OTP challenge table.

-- 1. Extend users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_verified BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS token_version INT NOT NULL DEFAULT 1;
ALTER TABLE users ADD COLUMN IF NOT EXISTS google_id VARCHAR(255) UNIQUE;
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- 2. Backfill existing accounts as verified so prior test/dev data remains active
UPDATE users SET is_verified = true, verified_at = now() WHERE is_verified = false;

-- 3. Create OTP challenges table
CREATE TABLE IF NOT EXISTS otps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    code_hash VARCHAR(64) NOT NULL,
    purpose VARCHAR(32) NOT NULL CHECK (purpose IN ('email_verification', 'password_reset')),
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index for rapid verification and cooldown checking
CREATE INDEX IF NOT EXISTS idx_otps_email_purpose_created 
    ON otps (email, purpose, created_at DESC);

