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

-- Create PostgreSQL roles for role-based access control
CREATE ROLE ledger_admin LOGIN PASSWORD 'admin_password';
CREATE ROLE ledger_viewer LOGIN PASSWORD 'viewer_password';

-- Grant connection permissions
GRANT CONNECT ON DATABASE railway TO ledger_admin, ledger_viewer;
GRANT USAGE ON SCHEMA public TO ledger_admin, ledger_viewer;

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
