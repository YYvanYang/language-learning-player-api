-- migrations/000001_create_users_table.down.sql

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS users;

-- Optional: Drop the extension if it's certain no other table needs it.
-- Usually, it's safe to leave it enabled.
-- DROP EXTENSION IF EXISTS "pgcrypto";