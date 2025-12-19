package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ledger-go-system/internal/auth"
)

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	Token   string `json:"token"`
	Role    string `json:"role"`
	Message string `json:"message,omitempty"`
}

type RefreshHandler struct {
	authManager    *auth.AuthManager
	userRepository *auth.UserRepository
}

func NewRefreshHandler(authManager *auth.AuthManager, userRepository *auth.UserRepository) *RefreshHandler {
	return &RefreshHandler{
		authManager:    authManager,
		userRepository: userRepository,
	}
}

// RefreshToken exchanges a refresh token for a new access token
func (h *RefreshHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.RefreshToken == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "refresh_token required"})
		return
	}

	// Verify the refresh token
	claims, err := h.authManager.VerifyRefreshToken(req.RefreshToken)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid or expired refresh token"})
		return
	}

	// Validate token in database
	if err := h.userRepository.ValidateRefreshToken(claims.UserID, req.RefreshToken); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "refresh token not found or revoked"})
		return
	}

	// Generate new access token
	newToken, err := h.authManager.GenerateToken(claims.Role, claims.UserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to generate token"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RefreshTokenResponse{
		Token:   newToken,
		Role:    claims.Role,
		Message: "token refreshed successfully",
	})
}

// RevokeRefreshToken revokes a refresh token
func (h *RefreshHandler) RevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "missing authorization header"})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid authorization header"})
		return
	}

	// Verify access token to get user ID
	claims, err := h.authManager.VerifyToken(parts[1])
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid token"})
		return
	}

	// Extract user ID from claims subject
	var userID int
	if _, err := fmt.Sscanf(claims.Subject, "%d", &userID); err != nil {
		userID = 0
	}

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.RefreshToken == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "refresh_token required"})
		return
	}

	// Revoke the refresh token
	if err := h.userRepository.RevokeRefreshToken(userID, req.RefreshToken); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to revoke token"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "refresh token revoked successfully",
	})
}
