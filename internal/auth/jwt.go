package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type AuthManager struct {
	secretKey string
}

// Users for demonstration (in production, use a real user database)
var ValidUsers = map[string]string{
	"admin":  "$2a$10$9x0.K5kpXZMqHt/tC0I8J.9u6L8sK8mD6vL9mP0qR2sT3uV4wX5yZ", // admin_password (hashed)
	"viewer": "$2a$10$8y1bL6lqYONnGuuUsD9hJ.8t5KcjL7nE5uK8lO9nQ1rS2tU3vW6xA",  // viewer_password (hashed)
}

func NewAuthManager(secretKey string) *AuthManager {
	return &AuthManager{secretKey: secretKey}
}

// GenerateToken creates a JWT token for the given role
func (am *AuthManager) GenerateToken(role string) (string, error) {
	claims := &Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(am.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

// VerifyToken validates and parses a JWT token
func (am *AuthManager) VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(am.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// VerifyCredentials checks username and password
func VerifyCredentials(username, password string) (string, error) {
	hashedPassword, exists := ValidUsers[username]
	if !exists {
		return "", fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	// Return the role based on username
	if username == "admin" {
		return "admin", nil
	}
	if username == "viewer" {
		return "viewer", nil
	}

	return "", fmt.Errorf("unknown role")
}

// HashPassword generates bcrypt hash of password (for reference)
// This is used to generate hashes for ValidUsers map
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
