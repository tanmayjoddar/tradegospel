package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"ledger-go-system/internal/auth"
)

type AuthHandler struct {
	authManager    *auth.AuthManager
	userRepository *auth.UserRepository
}

func NewAuthHandler(authManager *auth.AuthManager, userRepository *auth.UserRepository) *AuthHandler {
	return &AuthHandler{
		authManager:    authManager,
		userRepository: userRepository,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Role         string `json:"role"`
	ExpiresIn    int    `json:"expires_in"`
}

// Login authenticates user with username and password, returns JWT token and refresh token
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid request body"})
		return
	}

	if body.Username == "" || body.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "username and password required"})
		return
	}

	// Get user from database
	user, err := h.userRepository.GetUserByUsername(body.Username)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid credentials"})
		return
	}

	// Verify password
	if !h.userRepository.VerifyUserPassword(user, body.Password) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid credentials"})
		return
	}

	// Generate access token (1 hour)
	token, err := h.authManager.GenerateToken(user.Role, user.ID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to generate token"})
		return
	}

	// Generate refresh token (7 days)
	refreshToken, err := h.authManager.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to generate refresh token"})
		return
	}

	// Store refresh token in database
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := h.userRepository.CreateRefreshToken(user.ID, refreshToken, expiresAt); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to create refresh token"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Role:         user.Role,
		ExpiresIn:    3600, // 1 hour in seconds
	})
}

