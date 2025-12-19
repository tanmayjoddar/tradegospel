# Ledger System

Immutable ledger REST API with role-based access control.

## Setup

Prerequisites: Go 1.22+, PostgreSQL

Environment:

```env
DATABASE_URL=postgresql://user:pass@host:port/dbname
SERVER_PORT=8080
```

Install & Run:

```bash
go mod download
go generate ./...
go run cmd/server/main.go
```

## API

| Method | Endpoint     | Role         | Description  |
| ------ | ------------ | ------------ | ------------ |
| POST   | /ledger      | admin        | Create entry |
| GET    | /ledger      | admin/viewer | List entries |
| GET    | /ledger/{id} | admin/viewer | Get entry    |

Example:

```bash
curl -X POST http://localhost:8080/ledger \
  -H "Role: admin" \
  -H "Content-Type: application/json" \
  -d '{"amount": 100.50, "description": "Payment"}'
```

## Security

Application: No UPDATE/DELETE endpoints
Database: Roles prevent unauthorized modifications
Audit: All operations tracked

Role permissions:

```sql
GRANT INSERT, SELECT ON ledger TO ledger_admin;
GRANT SELECT ON ledger TO ledger_viewer;
REVOKE UPDATE, DELETE ON ledger FROM ledger_admin;
REVOKE UPDATE, DELETE ON ledger FROM ledger_viewer;
```

## Schema

Ledger: id, amount, description, created_at
AuditLog: id, ledger_id, actor, action, timestamp

## Tech

- Language: Go
- Database: PostgreSQL
- ORM: Prisma
- Auth: Header-based roles
