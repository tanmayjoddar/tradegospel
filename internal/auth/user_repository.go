package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *sql.DB
}

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

type RefreshToken struct {
	ID        int
	UserID    int
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetUserByUsername retrieves user from database by username
func (ur *UserRepository) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := ur.db.QueryRow(
		"SELECT id, username, password_hash, role, created_at FROM users WHERE username = $1",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

// VerifyUserPassword validates a user's password against the stored hash
func (ur *UserRepository) VerifyUserPassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// CreateRefreshToken stores a new refresh token in the database
func (ur *UserRepository) CreateRefreshToken(userID int, tokenString string, expiresAt time.Time) error {
	// Hash the token for storage (don't store plaintext tokens)
	tokenHash := sha256.Sum256([]byte(tokenString))
	tokenHashStr := hex.EncodeToString(tokenHash[:])

	_, err := ur.db.Exec(
		"INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)",
		userID, tokenHashStr, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

// ValidateRefreshToken checks if a refresh token is valid and not expired
func (ur *UserRepository) ValidateRefreshToken(userID int, tokenString string) error {
	// Hash the provided token
	tokenHash := sha256.Sum256([]byte(tokenString))
	tokenHashStr := hex.EncodeToString(tokenHash[:])

	var expiresAt time.Time
	err := ur.db.QueryRow(
		"SELECT expires_at FROM refresh_tokens WHERE user_id = $1 AND token_hash = $2",
		userID, tokenHashStr,
	).Scan(&expiresAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("refresh token not found or invalid")
	}
	if err != nil {
		return fmt.Errorf("failed to validate refresh token: %w", err)
	}

	if time.Now().After(expiresAt) {
		return fmt.Errorf("refresh token expired")
	}

	return nil
}

// RevokeRefreshToken removes a refresh token from the database
func (ur *UserRepository) RevokeRefreshToken(userID int, tokenString string) error {
	// Hash the provided token
	tokenHash := sha256.Sum256([]byte(tokenString))
	tokenHashStr := hex.EncodeToString(tokenHash[:])

	_, err := ur.db.Exec(
		"DELETE FROM refresh_tokens WHERE user_id = $1 AND token_hash = $2",
		userID, tokenHashStr,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// CleanupExpiredRefreshTokens removes expired tokens from database
func (ur *UserRepository) CleanupExpiredRefreshTokens() error {
	_, err := ur.db.Exec("DELETE FROM refresh_tokens WHERE expires_at < NOW()")
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	return nil
}

// CreateUser creates a new user in the database (for admin operations)
func (ur *UserRepository) CreateUser(username, password, role string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = ur.db.Exec(
		"INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3)",
		username, string(hashedPassword), role,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}
