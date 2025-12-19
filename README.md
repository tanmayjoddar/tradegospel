# Ledger System - Immutable REST API

A secure REST API with JWT authentication, refresh tokens, rate limiting, and database-stored credentials. Built to demonstrate immutability and production-grade security.

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
JWT_SECRET="your-super-secret-key-min-32-chars-change-production"
# Optional: HTTPS
# TLS_CERT="/path/to/cert.pem"
# TLS_KEY="/path/to/key.pem"
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

## Authentication with JWT & Refresh Tokens

All API endpoints require authentication. Access tokens expire in 1 hour - use refresh tokens to get new ones without logging in again.

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
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "role": "admin",
  "expires_in": 3600
}
```

**Response (401):**

```json
{
  "error": "invalid credentials"
}
```

### Step 2: Use Access Token

Add access token to every request:

```bash
-H "Authorization: Bearer <token>"
```

### Step 3: Refresh Token (When Expires)

**Request:**

```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "eyJhbGci..."}'
```

**Request Body:**

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "role": "admin",
  "message": "token refreshed successfully"
}
```

### Step 4: Logout (Revoke Refresh Token)

**Request:**

```bash
curl -X POST http://localhost:8080/auth/logout \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "eyJhbGci..."}'
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

REFRESH_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin_password"}' | jq -r '.refresh_token')

echo "Access Token: $TOKEN"
echo "Refresh Token: $REFRESH_TOKEN"

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

# 5. Refresh token (get new access token)
NEW_TOKEN=$(curl -s -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}" | jq -r '.token')

echo "New Access Token: $NEW_TOKEN"

# 6. Login as viewer
VIEWER_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "viewer", "password": "viewer_password"}' | jq -r '.token')

# 7. Viewer can read
curl -X GET http://localhost:8080/ledger \
  -H "Authorization: Bearer $VIEWER_TOKEN" | jq

# 8. Viewer CANNOT create (gets 403)
curl -X POST http://localhost:8080/ledger \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "description": "Test"}' | jq
```

---

## Features

✅ **JWT Authentication** - Secure token-based auth with 1-hour expiration
✅ **Refresh Tokens** - 7-day refresh tokens for seamless UX (no repeated logins)
✅ **Database Credentials** - User passwords stored securely in database with bcrypt
✅ **Rate Limiting** - 60 requests per minute per IP to prevent abuse
✅ **HTTPS Support** - Optional TLS/SSL configuration for production
✅ **Role-Based Access** - Admin creates, Viewer reads only
✅ **Immutable Ledger** - No UPDATE/DELETE allowed at app or DB level
✅ **Audit Log** - Track all INSERT operations with actor and timestamp
✅ **Bcrypt Passwords** - Secure password hashing and verification

---

## Security

### Roles

| Role   | Create | Read | Update | Delete |
| ------ | ------ | ---- | ------ | ------ |
| Admin  | ✅     | ✅   | ❌     | ❌     |
| Viewer | ❌     | ✅   | ❌     | ❌     |

### How It Works

1. **JWT prevents role spoofing** - Viewer cannot claim to be admin (signature verification fails)
2. **Bcrypt password hashing** - Passwords stored securely, never in plaintext
3. **Database enforces immutability** - Even if app is compromised, DB blocks UPDATE/DELETE
4. **Refresh token rotation** - Tokens stored in database with hash verification
5. **Rate limiting** - Prevents brute force and DoS attacks
6. **HTTPS ready** - TLS configuration available for encrypted communication

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

### Rate Limited

```json
{
  "error": "rate limit exceeded - try again in 1 minute"
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

### users table

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### refresh_tokens table

```sql
CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### rate_limit_log table

```sql
CREATE TABLE rate_limit_log (
    id SERIAL PRIMARY KEY,
    ip_address VARCHAR(45) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    request_count INTEGER DEFAULT 1,
    window_start TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ip_address, endpoint, window_start)
);
```

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

---

## Project Structure

```
.
├── cmd/server/main.go              - Server setup, TLS, rate limiting
├── internal/
│   ├── auth/
│   │   ├── jwt.go                  - JWT & refresh token generation
│   │   └── user_repository.go      - User & token database queries
│   ├── handler/
│   │   ├── auth_handler.go         - Login endpoint
│   │   ├── refresh_handler.go      - Token refresh & logout
│   │   └── ledger_handler.go       - Ledger CRUD endpoints
│   ├── middleware/
│   │   ├── role.go                 - JWT verification & authorization
│   │   └── rate_limit.go           - Request rate limiting
│   ├── repository/ledger_repo.go   - Database queries
│   └── db/postgres.go              - DB connection
├── database/schema.sql             - Database setup & users
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

- Access token expires after 1 hour
- Use refresh token to get a new one: `POST /auth/refresh`

**"rate limit exceeded"**

- Too many requests from this IP
- Wait 1 minute and try again

**"missing authorization header"**

- Add header: `-H "Authorization: Bearer <token>"`

---


## How Immutability Works

**You cannot UPDATE or DELETE entries because:**

1. **App Level** - No UPDATE or DELETE endpoints exist
2. **Database Level** - PostgreSQL roles prevent UPDATE/DELETE even if you try SQL directly

This is why immutability is enforced at two layers - if one fails, the other protects the data.

---
