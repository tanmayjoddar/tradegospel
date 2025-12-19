package middleware

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	requestsPerMinute = 60
	windowDuration    = 1 * time.Minute
)

type RateLimiter struct {
	db *sql.DB
}

func NewRateLimiter(db *sql.DB) *RateLimiter {
	return &RateLimiter{db: db}
}

// getClientIP extracts the client's IP address from the request
func getClientIP(r *http.Request) string {
	// Try X-Forwarded-For first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Try X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// RateLimit middleware enforces request rate limiting per IP and endpoint
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		endpoint := r.Method + " " + r.RequestURI

		// Check and update rate limit
		allowed, err := rl.checkRateLimit(clientIP, endpoint)
		if err != nil {
			// Log error but allow request to continue (fail open)
			fmt.Printf("Rate limit check error: %v\n", err)
		}

		if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded - try again in 1 minute"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// checkRateLimit checks if the request is within the rate limit
func (rl *RateLimiter) checkRateLimit(ip, endpoint string) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-windowDuration)

	// Use a transaction to ensure atomicity
	tx, err := rl.db.Begin()
	if err != nil {
		return true, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clean up old entries outside the window
	_, err = tx.Exec(
		"DELETE FROM rate_limit_log WHERE ip_address = $1 AND endpoint = $2 AND window_start < $3",
		ip, endpoint, windowStart,
	)
	if err != nil {
		return true, fmt.Errorf("failed to cleanup old entries: %w", err)
	}

	// Try to increment existing entry
	res, err := tx.Exec(
		"UPDATE rate_limit_log SET request_count = request_count + 1 WHERE ip_address = $1 AND endpoint = $2 AND window_start > $3 RETURNING request_count",
		ip, endpoint, windowStart,
	)
	if err != nil {
		return true, fmt.Errorf("failed to update rate limit: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return true, fmt.Errorf("failed to get rows affected: %w", err)
	}

	var currentCount int

	if rowsAffected == 0 {
		// Insert new entry
		err = tx.QueryRow(
			"INSERT INTO rate_limit_log (ip_address, endpoint, window_start) VALUES ($1, $2, $3) RETURNING request_count",
			ip, endpoint, now,
		).Scan(&currentCount)
		if err != nil {
			return true, fmt.Errorf("failed to insert rate limit entry: %w", err)
		}
		currentCount = 1
	} else {
		// Get updated count
		err = tx.QueryRow(
			"SELECT request_count FROM rate_limit_log WHERE ip_address = $1 AND endpoint = $2 ORDER BY window_start DESC LIMIT 1",
			ip, endpoint,
		).Scan(&currentCount)
		if err != nil && err != sql.ErrNoRows {
			return true, fmt.Errorf("failed to get rate limit count: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return true, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Check if exceeded
	if currentCount > requestsPerMinute {
		return false, nil
	}

	return true, nil
}

// CleanupExpiredLogs removes old rate limit logs (runs periodically)
func (rl *RateLimiter) CleanupExpiredLogs() error {
	windowStart := time.Now().Add(-24 * time.Hour)
	_, err := rl.db.Exec(
		"DELETE FROM rate_limit_log WHERE window_start < $1",
		windowStart,
	)
	if err != nil {
		return fmt.Errorf("failed to cleanup rate limit logs: %w", err)
	}
	return nil
}
