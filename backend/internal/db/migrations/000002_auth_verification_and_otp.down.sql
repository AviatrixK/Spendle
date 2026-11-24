-- Migration: 000002_auth_verification_and_otp.down.sql
-- Description: Rollback email verification, session invalidation, Google OAuth, and OTP table.

DROP TABLE IF EXISTS otps CASCADE;

ALTER TABLE users DROP COLUMN IF EXISTS is_verified;
ALTER TABLE users DROP COLUMN IF EXISTS verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS token_version;
ALTER TABLE users DROP COLUMN IF EXISTS google_id;
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;

