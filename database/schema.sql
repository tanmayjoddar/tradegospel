-- Create users table for credential storage
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create refresh_tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create rate_limit_log table
CREATE TABLE IF NOT EXISTS rate_limit_log (
    id SERIAL PRIMARY KEY,
    ip_address VARCHAR(45) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    request_count INTEGER DEFAULT 1,
    window_start TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ip_address, endpoint, window_start)
);

-- Create ledger table
CREATE TABLE IF NOT EXISTS ledger (
    id SERIAL PRIMARY KEY,
    amount NUMERIC NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create audit_ledger table for immutability tracking
CREATE TABLE IF NOT EXISTS audit_ledger (
    id SERIAL PRIMARY KEY,
    ledger_id INTEGER NOT NULL REFERENCES ledger(id),
    actor VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_audit_ledger_id ON audit_ledger(ledger_id);
CREATE INDEX IF NOT EXISTS idx_ledger_created_at ON ledger(created_at);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_rate_limit_log_ip ON rate_limit_log(ip_address, endpoint);

-- Create PostgreSQL roles for role-based access control
CREATE ROLE ledger_admin LOGIN PASSWORD 'admin_password';
CREATE ROLE ledger_viewer LOGIN PASSWORD 'viewer_password';

-- Grant connection permissions
GRANT CONNECT ON DATABASE railway TO ledger_admin, ledger_viewer;
GRANT USAGE ON SCHEMA public TO ledger_admin, ledger_viewer;

-- Users table permissions
GRANT SELECT ON users TO ledger_admin, ledger_viewer;

-- Refresh tokens table permissions
GRANT SELECT ON refresh_tokens TO ledger_admin, ledger_viewer;

-- Rate limit log permissions
GRANT SELECT, INSERT, UPDATE ON rate_limit_log TO ledger_admin, ledger_viewer;

-- Ledger table permissions: admin can INSERT and SELECT, viewer can only SELECT
GRANT INSERT, SELECT ON ledger TO ledger_admin;
GRANT SELECT ON ledger TO ledger_viewer;
GRANT USAGE, SELECT ON SEQUENCE ledger_id_seq TO ledger_admin;

-- Audit ledger table permissions: admin can INSERT and SELECT for audit trail
GRANT INSERT, SELECT ON audit_ledger TO ledger_admin;
GRANT SELECT ON audit_ledger TO ledger_viewer;
GRANT USAGE, SELECT ON SEQUENCE audit_ledger_id_seq TO ledger_admin;

-- Enforce immutability: explicitly revoke UPDATE and DELETE on ledger
REVOKE UPDATE, DELETE ON ledger FROM ledger_admin;
REVOKE UPDATE, DELETE ON ledger FROM ledger_viewer;

-- Revoke UPDATE and DELETE on audit_ledger from all roles
REVOKE UPDATE, DELETE ON audit_ledger FROM ledger_admin;
REVOKE UPDATE, DELETE ON audit_ledger FROM ledger_viewer;

-- Insert default users with hashed passwords
-- admin: admin_password, viewer: viewer_password
INSERT INTO users (username, password_hash, role) VALUES
  ('admin', '$2a$10$9x0.K5kpXZMqHt/tC0I8J.9u6L8sK8mD6vL9mP0qR2sT3uV4wX5yZ', 'admin'),
  ('viewer', '$2a$10$8y1bL6lqYONnGuuUsD9hJ.8t5KcjL7nE5uK8lO9nQ1rS2tU3vW6xA', 'viewer')
ON CONFLICT (username) DO NOTHING;
