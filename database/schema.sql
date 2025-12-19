CREATE TABLE ledger (
    id SERIAL PRIMARY KEY,
    amount NUMERIC(10,2) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE ROLE ledger_admin LOGIN PASSWORD 'admin_password';
CREATE ROLE ledger_viewer LOGIN PASSWORD 'viewer_password';

GRANT CONNECT ON DATABASE railway TO ledger_admin, ledger_viewer;
GRANT USAGE ON SCHEMA public TO ledger_admin, ledger_viewer;

GRANT INSERT, SELECT ON ledger TO ledger_admin;
GRANT SELECT ON ledger TO ledger_viewer;

REVOKE UPDATE, DELETE ON ledger FROM ledger_admin;
REVOKE UPDATE, DELETE ON ledger FROM ledger_viewer;
