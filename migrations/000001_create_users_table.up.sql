-- migrations/000001_create_users_table.up.sql

-- Enable UUID generation if not already enabled
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create the users table based on domain model and Section 7 DDL
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    password_hash VARCHAR(255) NULL, -- Nullable for external auth providers
    name VARCHAR(100),
    google_id VARCHAR(255) UNIQUE NULL,
    auth_provider VARCHAR(50) NOT NULL DEFAULT 'local', -- 'local', 'google' etc.
    profile_image_url VARCHAR(1024) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Add indexes for faster lookups
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_google_id ON users(google_id) WHERE google_id IS NOT NULL;
CREATE INDEX idx_users_auth_provider ON users(auth_provider);

-- Optional: Add a trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();