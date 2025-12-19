package middleware

import (
	"context"
	"net/http"
	"strings"

	"ledger-go-system/internal/auth"
)

// RoleKey is used for context values
type contextKey string

const RoleKey contextKey = "role"

// JWTMiddleware validates JWT tokens and extracts role information
func JWTMiddleware(authManager *auth.AuthManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Bearer token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Verify and parse the token
		claims, err := authManager.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Store role in context for downstream handlers
		ctx := context.WithValue(r.Context(), RoleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware ensures the request has a specific role
func RequireRole(requiredRole string, authManager *auth.AuthManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Bearer token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Verify and parse the token
		claims, err := authManager.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Check if role matches required role
		if claims.Role != requiredRole {
			http.Error(w, `{"error":"forbidden - insufficient permissions"}`, http.StatusForbidden)
			return
		}

		// Store role in context for downstream handlers
		ctx := context.WithValue(r.Context(), RoleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AllowRoles middleware allows multiple roles
func AllowRoles(allowedRoles []string, authManager *auth.AuthManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Bearer token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Verify and parse the token
		claims, err := authManager.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Check if role is in allowed list
		roleAllowed := false
		for _, role := range allowedRoles {
			if claims.Role == role {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			http.Error(w, `{"error":"forbidden - insufficient permissions"}`, http.StatusForbidden)
			return
		}

		// Store role in context for downstream handlers
		ctx := context.WithValue(r.Context(), RoleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRoleFromContext retrieves the role from request context
func GetRoleFromContext(r *http.Request) string {
	role, ok := r.Context().Value(RoleKey).(string)
	if !ok {
		return "unknown"
	}
	return role
}
