-- migrations/000004_create_refresh_tokens.up.sql

CREATE TABLE refresh_tokens (
    -- id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Use hash as PK? No, better to have separate ID
    -- Store the SHA-256 hash of the refresh token value
    -- Use TEXT as hash length is fixed (64 hex chars) but TEXT is simpler. Or VARCHAR(64).
    token_hash TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    -- revoked_at TIMESTAMPTZ NULL -- Optional: Add if explicit revocation tracking is needed besides deletion
);

-- Index for faster lookup by user_id (e.g., for deleting all tokens for a user)
CREATE INDEX idx_refreshtokens_user_id ON refresh_tokens(user_id);

-- Index on expires_at for potential cleanup jobs
CREATE INDEX idx_refreshtokens_expires_at ON refresh_tokens(expires_at);