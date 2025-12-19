# Immutable Ledger System — Production-Ready REST API

A secure, enterprise-grade REST API demonstrating advanced backend development practices including JWT authentication with refresh tokens, role-based access control, rate limiting, and cryptographic security.

**Live Demo:** [GitHub Repository](https://github.com/TJ456/tradegospel)

---

## 🎯 Key Features

| Feature                 | Implementation                                     |
| ----------------------- | -------------------------------------------------- |
| **Authentication**      | JWT (HMAC-SHA256) with 1-hour access tokens        |
| **Credential Security** | Bcrypt password hashing stored in PostgreSQL       |
| **Long Sessions**       | 7-day refresh tokens with database validation      |
| **Rate Limiting**       | 60 requests/minute per IP (DoS prevention)         |
| **Immutability**        | Defense-in-depth: app-level + DB-level enforcement |
| **Authorization**       | Role-based access control (Admin/Viewer)           |
| **Audit Trail**         | Complete logging of all operations                 |
| **HTTPS Ready**         | Optional TLS/SSL configuration                     |
| **Atomic Transactions** | Consistent ledger + audit log writes               |

---

## ⚡ Quick Start (5 minutes)

### Prerequisites

- **Go 1.22+** - [Download](https://golang.org/dl/)
- **PostgreSQL 12+** - Local OR Railway hosted

### 1️⃣ Clone Repository

```bash
git clone https://github.com/TJ456/tradegospel.git
cd tradegospel
```

### 2️⃣ Configure Database

**Option A: Local PostgreSQL**

```bash
# Create database and run schema
psql -U postgres -d postgres -f database/schema.sql
```

**Option B: Railway Hosted (Recommended)**

```bash
# Copy your Railway DATABASE_URL from dashboard
cp .env.example .env

# Edit .env with your Railway connection:
# DATABASE_URL="postgresql://user:pass@containers.railway.app:7070/railway"
```

### 3️⃣ Setup Environment

```bash
# Copy environment template
cp .env.example .env

# Update .env with your database and JWT secret
cat .env
```

**Required environment variables:**

```env
DATABASE_URL="postgresql://user:password@host:5432/database"
SERVER_PORT=8080
JWT_SECRET="change-this-to-random-32-character-secret"
```

### 4️⃣ Start Server

```bash
go mod download
go run cmd/server/main.go
```

✅ Server running on `http://localhost:8080`

---

## 🔐 Authentication Flow (For Reviewers)

### Step 1: Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin_password"
  }'
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "role": "admin",
  "expires_in": 3600
}
```

### Step 2: Use Access Token

All requests require `Authorization` header:

```bash
curl -X GET http://localhost:8080/ledger \
  -H "Authorization: Bearer <your_token>"
```

### Step 3: Refresh Token (When Expired)

```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "<your_refresh_token>"}'
```

---

## 📊 Complete API Reference

### Authentication Endpoints

#### **POST /auth/login** — Get access & refresh tokens

```bash
REQUEST:
{
  "username": "admin",
  "password": "admin_password"
}

RESPONSE (200):
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "role": "admin",
  "expires_in": 3600
}

RESPONSE (401):
{
  "error": "invalid credentials"
}
```

#### **POST /auth/refresh** — Get new access token

```bash
REQUEST:
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}

RESPONSE (200):
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "role": "admin",
  "message": "token refreshed successfully"
}
```

#### **POST /auth/logout** — Revoke refresh token

```bash
REQUEST:
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}

RESPONSE (200):
{
  "message": "refresh token revoked successfully"
}
```

### Ledger Endpoints

#### **POST /ledger** — Create entry (Admin only)

```bash
REQUEST:
{
  "amount": 150.75,
  "description": "Monthly salary payment"
}

RESPONSE (201):
{
  "status": "created"
}

RESPONSE (403 - if viewer):
{
  "error": "forbidden - insufficient permissions"
}
```

#### **GET /ledger** — List all entries (Admin & Viewer)

```bash
RESPONSE (200):
[
  {
    "id": 1,
    "amount": 150.75,
    "description": "Monthly salary",
    "created_at": "2025-12-19T10:30:45Z"
  }
]
```

#### **GET /ledger/{id}** — Get single entry (Admin & Viewer)

```bash
RESPONSE (200):
{
  "id": 1,
  "amount": 150.75,
  "description": "Monthly salary",
  "created_at": "2025-12-19T10:30:45Z"
}

RESPONSE (404):
{
  "error": "ledger entry not found"
}
```

---

## 🧪 Test the Complete Flow

Save this as `test.sh` and run:

```bash
#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Testing Immutable Ledger API ===${NC}\n"

# 1. Login as admin
echo -e "${GREEN}1. Admin Login${NC}"
ADMIN_RESPONSE=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin_password"}')

ADMIN_TOKEN=$(echo $ADMIN_RESPONSE | jq -r '.token')
ADMIN_REFRESH=$(echo $ADMIN_RESPONSE | jq -r '.refresh_token')
echo "✓ Admin Token: ${ADMIN_TOKEN:0:20}..."
echo "✓ Refresh Token: ${ADMIN_REFRESH:0:20}...\n"

# 2. Create ledger entry
echo -e "${GREEN}2. Create Ledger Entry${NC}"
curl -s -X POST http://localhost:8080/ledger \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount":1500.50,"description":"Q4 Revenue"}' | jq .
echo ""

# 3. List entries
echo -e "${GREEN}3. List All Entries${NC}"
curl -s -X GET http://localhost:8080/ledger \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
echo ""

# 4. Get single entry
echo -e "${GREEN}4. Get Entry by ID${NC}"
curl -s -X GET http://localhost:8080/ledger/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
echo ""

# 5. Test viewer cannot create
echo -e "${GREEN}5. Viewer Trying to Create (Should Fail)${NC}"
VIEWER_RESPONSE=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"viewer","password":"viewer_password"}')

VIEWER_TOKEN=$(echo $VIEWER_RESPONSE | jq -r '.token')
echo "Viewer Token: ${VIEWER_TOKEN:0:20}..."

curl -s -X POST http://localhost:8080/ledger \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount":100,"description":"Unauthorized"}' | jq .
echo ""

# 6. Test rate limiting
echo -e "${GREEN}6. Rate Limiting Test${NC}"
for i in {1..65}; do
  curl -s -X GET http://localhost:8080/ledger \
    -H "Authorization: Bearer $ADMIN_TOKEN" > /dev/null
done
echo "✓ Made 65 requests (limit is 60/min)"
curl -s -X GET http://localhost:8080/ledger \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.error'
echo ""

# 7. Refresh token
echo -e "${GREEN}7. Refresh Access Token${NC}"
curl -s -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$ADMIN_REFRESH\"}" | jq .
echo ""

echo -e "${GREEN}✅ All tests completed!${NC}"
```

Run it:

```bash
chmod +x test.sh
./test.sh
```

---

## 🏗️ Architecture

### Project Structure

```
tradegospel/
├── cmd/server/main.go                    # Server entry point with TLS & rate limiting
├── internal/
│   ├── auth/
│   │   ├── jwt.go                        # JWT generation & verification
│   │   └── user_repository.go            # Database user & token queries
│   ├── handler/
│   │   ├── auth_handler.go               # Login with credential verification
│   │   ├── refresh_handler.go            # Token refresh & logout
│   │   └── ledger_handler.go             # Immutable ledger CRUD
│   ├── middleware/
│   │   ├── role.go                       # JWT validation & RBAC
│   │   └── rate_limit.go                 # Per-IP request rate limiting
│   ├── repository/
│   │   └── ledger_repository.go          # Database queries
│   └── db/
│       └── postgres.go                   # Connection pooling
├── database/
│   └── schema.sql                        # Complete schema with users & tokens
├── .env.example                          # Environment template
└── README.md                             # This file
```

### Security Architecture

```
┌─────────────────────────────────────────────────┐
│                  Client Request                  │
└────────────────┬────────────────────────────────┘
                 │
         ┌───────▼────────┐
         │ Rate Limiter   │  60 req/min/IP
         └───────┬────────┘
                 │
         ┌───────▼────────┐
         │ JWT Middleware │  Verify signature
         └───────┬────────┘
                 │
         ┌───────▼────────┐
         │ RBAC Middleware│  Check role
         └───────┬────────┘
                 │
         ┌───────▼────────┐
         │ Handler Logic  │  Process request
         └───────┬────────┘
                 │
         ┌───────▼────────┐
         │  Transaction   │  Ledger + Audit
         └───────┬────────┘
                 │
         ┌───────▼────────┐
         │ PostgreSQL     │  REVOKE UPDATE/DELETE
         │ Enforces       │  Role-based permissions
         └────────────────┘
```

---

## 🔒 Security Deep Dive

### Why Viewer Cannot Become Admin

1. **JWT Signature Verification**

   - Token signed with server's secret key (HMAC-SHA256)
   - Viewer cannot forge admin token without secret
   - Any modification invalidates signature

2. **Database Role-Based Permissions**

   - Admin role has: INSERT, SELECT on ledger
   - Viewer role has: SELECT only
   - Even if app is compromised: `UPDATE ledger SET ...` returns permission denied

3. **Refresh Token Validation**
   - Tokens stored in database with SHA256 hash
   - Cannot be modified in flight
   - Expiration checked on every refresh request

### How Immutability Works

**App Level:** No DELETE/UPDATE endpoints exist
**DB Level:** PostgreSQL enforces via roles

```sql
REVOKE UPDATE, DELETE ON ledger FROM ledger_admin;
REVOKE UPDATE, DELETE ON ledger FROM ledger_viewer;
```

---

## ✅ Production Checklist

- [x] Passwords stored securely (bcrypt hashing)
- [x] JWT with short expiration (1 hour)
- [x] Refresh tokens with longer expiration (7 days)
- [x] Rate limiting prevents brute force
- [x] Database enforces immutability
- [x] HTTPS/TLS support
- [x] Atomic transactions
- [x] Audit logging
- [x] Error handling
- [x] Input validation

---

## 📚 Technical Highlights

**JWT Implementation**

- HMAC-SHA256 signing algorithm
- Subject claim stores user ID
- Role claim extracted on every request
- Expiration validated automatically

**Password Security**

- Bcrypt cost factor: 10
- Stored in users table, never in memory
- Compared using time-constant function

**Rate Limiting**

- Per-IP, per-endpoint tracking
- 1-minute rolling windows
- Automatic cleanup of old entries
- Database-backed (survives server restart)

**Refresh Token Flow**

- Access token: 1 hour (stateless)
- Refresh token: 7 days (stateful, in DB)
- Token hash stored, not plaintext
- Can be revoked immediately

---

## 🐛 Troubleshooting

**"Connection refused"**

```bash
# Check if PostgreSQL is running
psql -U postgres -c "SELECT 1;"

# Or check Railway connection string
echo $DATABASE_URL
```

**"Invalid or expired token"**

```bash
# Access tokens expire after 1 hour
# Use refresh token to get new one:
curl -X POST http://localhost:8080/auth/refresh \
  -d '{"refresh_token": "..."}'
```

**"Rate limit exceeded"**

```bash
# 60 requests per minute per IP
# Wait 1 minute and retry
```

**"Permission denied" (database)**

```bash
# Schema not applied or wrong permissions
psql -U postgres -d ledger_db -c "GRANT INSERT, SELECT ON ledger TO ledger_admin;"
```

---

## 📖 Additional Documentation

- **[SETUP_GUIDE.md](SETUP_GUIDE.md)** — Detailed installation & deployment
- **[DESIGN.md](DESIGN.md)** — Architecture & design decisions
- **[IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)** — Security implementation
- **[FINAL_CHECKLIST.md](FINAL_CHECKLIST.md)** — Requirements verification

---

## 🚀 Next Steps

### For Local Development

```bash
go run cmd/server/main.go
# Server ready at http://localhost:8080
```

### For Railway Deployment

```bash
# 1. Connect GitHub repository to Railway
# 2. Set environment variables in Railway dashboard
# 3. Railway auto-deploys on git push
```

### For Production

```bash
# 1. Generate proper TLS certificates (Let's Encrypt)
# 2. Set strong JWT_SECRET (32+ characters)
# 3. Configure DATABASE_URL with SSL
# 4. Enable password reset mechanism
# 5. Setup monitoring & logging
```

---

## 💡 Key Takeaways

This project demonstrates:

- ✅ Secure credential handling (bcrypt + JWT)
- ✅ Defense-in-depth security (app + DB)
- ✅ Stateless + stateful token patterns
- ✅ Production-grade error handling
- ✅ Rate limiting & DoS prevention
- ✅ Role-based access control
- ✅ Database transaction atomicity
- ✅ Clean Go architecture

---

**Questions?** Feel free to open an issue or reach out directly.

Happy reviewing! 🎉
