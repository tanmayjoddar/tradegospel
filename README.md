# Ledger System - Immutable REST API

A simple REST API with JWT authentication and role-based access control. Built to demonstrate immutability and security.

## Setup

### Prerequisites

- Go 1.22+
- PostgreSQL 12+

### Installation

1. **Clone and setup**

```bash
cp .env.example .env
```

2. **Configure `.env`**

```env
DATABASE_URL="postgresql://ledger_admin:admin_password@localhost:5432/ledger_db"
SERVER_PORT=8080
JWT_SECRET="your-secret-key"
```

3. **Create database**

```bash
psql -U postgres -d postgres -f database/schema.sql
```

4. **Run server**

```bash
go mod download
go run cmd/server/main.go
```

Server runs on `http://localhost:8080`

---

## Authentication with JWT

You need a **JWT token** to access all API endpoints.

### Step 1: Login

**Credentials:**

- Admin: `admin` / `admin_password`
- Viewer: `viewer` / `viewer_password`

**Request:**

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin_password"}'
```

**Request Body:**

```json
{
  "username": "admin",
  "password": "admin_password"
}
```

**Response (200):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "role": "admin"
}
```

**Response (401):**

```json
{
  "error": "invalid credentials"
}
```

### Step 2: Use Token

Add token to every request in the `Authorization` header:

```bash
-H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## API Endpoints

### 1. Create Entry (Admin Only)

**POST** `/ledger`

**Headers:**

```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**

```json
{
  "amount": 150.75,
  "description": "Monthly salary"
}
```

**Response (201):**

```json
{
  "status": "created"
}
```

**Response (403 - if viewer):**

```json
{
  "error": "forbidden - insufficient permissions"
}
```

---

### 2. List All Entries

**GET** `/ledger`

**Headers:**

```
Authorization: Bearer <token>
```

**Response (200):**

```json
[
  {
    "id": 1,
    "amount": 150.75,
    "description": "Monthly salary",
    "created_at": "2025-12-19T10:30:45Z"
  },
  {
    "id": 2,
    "amount": 200.0,
    "description": "Expense reimbursement",
    "created_at": "2025-12-19T11:15:30Z"
  }
]
```

---

### 3. Get Single Entry

**GET** `/ledger/{id}`

**Headers:**

```
Authorization: Bearer <token>
```

**Example:** `GET /ledger/1`

**Response (200):**

```json
{
  "id": 1,
  "amount": 150.75,
  "description": "Monthly salary",
  "created_at": "2025-12-19T10:30:45Z"
}
```

**Response (404):**

```json
{
  "error": "ledger entry not found"
}
```

---

## Quick Test

```bash
# 1. Login as admin
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin_password"}' | jq -r '.token')

echo "Token: $TOKEN"

# 2. Create entry
curl -X POST http://localhost:8080/ledger \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount": 150.75, "description": "Monthly salary"}' | jq

# 3. List all
curl -X GET http://localhost:8080/ledger \
  -H "Authorization: Bearer $TOKEN" | jq

# 4. Get entry by ID
curl -X GET http://localhost:8080/ledger/1 \
  -H "Authorization: Bearer $TOKEN" | jq

# 5. Login as viewer
VIEWER_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "viewer", "password": "viewer_password"}' | jq -r '.token')

# 6. Viewer can read
curl -X GET http://localhost:8080/ledger \
  -H "Authorization: Bearer $VIEWER_TOKEN" | jq

# 7. Viewer CANNOT create (gets 403)
curl -X POST http://localhost:8080/ledger \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "description": "Test"}' | jq
```

---

## Features

✅ JWT Authentication - Secure token-based auth
✅ Role-Based Access - Admin creates, Viewer reads only
✅ Immutable Ledger - No UPDATE/DELETE allowed
✅ Audit Log - Track all INSERT operations
✅ Bcrypt Passwords - Secure password storage
✅ Database Protection - PostgreSQL enforces immutability

---

## Security

### Roles

| Role   | Create | Read | Update | Delete |
| ------ | ------ | ---- | ------ | ------ |
| Admin  | ✅     | ✅   | ❌     | ❌     |
| Viewer | ❌     | ✅   | ❌     | ❌     |

### How It Works

1. **JWT prevents role spoofing** - Viewer cannot claim to be admin (signature won't match)
2. **Database prevents updates** - Even if app is compromised, DB blocks UPDATE/DELETE
3. **Audit trail** - Every action is logged with who did it

### Test Database Protection

```bash
# Try to update directly (will fail)
psql -U ledger_admin -d ledger_db -c "UPDATE ledger SET amount = 999 WHERE id = 1;"
# Error: permission denied for relation ledger

# Try to delete (will fail)
psql -U ledger_admin -d ledger_db -c "DELETE FROM ledger WHERE id = 1;"
# Error: permission denied for relation ledger
```

---

## Error Responses

### Missing Token

```json
{
  "error": "missing authorization header"
}
```

### Invalid Token

```json
{
  "error": "invalid or expired token"
}
```

### Wrong Role

```json
{
  "error": "forbidden - insufficient permissions"
}
```

### Invalid Amount

```json
{
  "error": "amount must be positive"
}
```

### Not Found

```json
{
  "error": "ledger entry not found"
}
```

---

## Database Schema

### ledger table

```sql
CREATE TABLE ledger (
    id SERIAL PRIMARY KEY,
    amount NUMERIC NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### audit_ledger table

```sql
CREATE TABLE audit_ledger (
    id SERIAL PRIMARY KEY,
    ledger_id INTEGER NOT NULL REFERENCES ledger(id),
    actor VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Roles

```sql
-- Admin: Can INSERT and SELECT
CREATE ROLE ledger_admin LOGIN PASSWORD 'admin_password';
GRANT INSERT, SELECT ON ledger TO ledger_admin;

-- Viewer: Can only SELECT
CREATE ROLE ledger_viewer LOGIN PASSWORD 'viewer_password';
GRANT SELECT ON ledger TO ledger_viewer;

-- Both: Cannot UPDATE or DELETE
REVOKE UPDATE, DELETE ON ledger FROM ledger_admin;
REVOKE UPDATE, DELETE ON ledger FROM ledger_viewer;
```

---

## Project Structure

```
.
├── cmd/server/main.go              - Server setup & routes
├── internal/
│   ├── auth/jwt.go                 - JWT & credentials
│   ├── handler/
│   │   ├── auth_handler.go         - Login endpoint
│   │   └── ledger_handler.go       - Ledger endpoints
│   ├── middleware/role.go          - JWT verification
│   ├── repository/ledger_repo.go   - Database queries
│   └── db/postgres.go              - DB connection
├── database/schema.sql             - Database setup
├── .env.example                    - Environment variables
└── go.mod                          - Go dependencies
```

---

## Troubleshooting

**"connection refused"**

- PostgreSQL not running
- Start: `brew services start postgresql` (macOS)

**"permission denied"**

- Run schema setup: `psql -U postgres -d postgres -f database/schema.sql`

**"invalid or expired token"**

- Token expires after 24 hours
- Login again to get new token

**"missing authorization header"**

- Add header: `-H "Authorization: Bearer <token>"`

---

## Production Checklist

- [ ] Change `JWT_SECRET` to strong random key
- [ ] Use HTTPS only
- [ ] Store passwords in real database
- [ ] Implement refresh tokens
- [ ] Add rate limiting
- [ ] Enable PostgreSQL SSL
- [ ] Setup backups
- [ ] Add request logging

---

## How Immutability Works

**You cannot UPDATE or DELETE entries because:**

1. **App Level** - No UPDATE or DELETE endpoints exist
2. **Database Level** - PostgreSQL roles prevent UPDATE/DELETE even if you try SQL directly

This is why immutability is enforced at two layers - if one fails, the other protects the data.

---

**See SETUP_GUIDE.md and DESIGN.md for more details**
