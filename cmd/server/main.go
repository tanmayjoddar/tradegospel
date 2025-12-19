package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"ledger-go-system/internal/auth"
	"ledger-go-system/internal/db"
	"ledger-go-system/internal/handler"
	"ledger-go-system/internal/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production" // Change in .env for production
	}

	tlsCert := os.Getenv("TLS_CERT")
	tlsKey := os.Getenv("TLS_KEY")

	conn, err := db.New(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	authManager := auth.NewAuthManager(jwtSecret)
	userRepository := auth.NewUserRepository(conn)
	ledgerHandler := handler.NewLedgerHandler(conn)
	authHandler := handler.NewAuthHandler(authManager, userRepository)
	refreshHandler := handler.NewRefreshHandler(authManager, userRepository)
	rateLimiter := middleware.NewRateLimiter(conn)

	// Cleanup expired tokens periodically
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := userRepository.CleanupExpiredRefreshTokens(); err != nil {
				log.Printf("Failed to cleanup expired tokens: %v", err)
			}
			if err := rateLimiter.CleanupExpiredLogs(); err != nil {
				log.Printf("Failed to cleanup rate limit logs: %v", err)
			}
		}
	}()

	mux := http.NewServeMux()

	// Apply rate limiting to all endpoints
	rateLimitedMux := rateLimiter.RateLimit(mux)

	// Public endpoints (no auth required)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", refreshHandler.RefreshToken)
	mux.HandleFunc("POST /auth/logout", refreshHandler.RevokeRefreshToken)

	// Protected endpoints (require JWT)
	// Admin only: POST /ledger
	mux.Handle("POST /ledger", middleware.RequireRole("admin", authManager, http.HandlerFunc(ledgerHandler.Create)))

	// Both admin and viewer: GET /ledger, GET /ledger/{id}
	mux.Handle("GET /ledger", middleware.AllowRoles([]string{"admin", "viewer"}, authManager, http.HandlerFunc(ledgerHandler.List)))
	mux.Handle("GET /ledger/", middleware.AllowRoles([]string{"admin", "viewer"}, authManager, http.HandlerFunc(ledgerHandler.GetByID)))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      rateLimitedMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server with HTTPS if certificates are provided
	if tlsCert != "" && tlsKey != "" {
		log.Printf("Server running on https://localhost:%s", port)
		if err := server.ListenAndServeTLS(tlsCert, tlsKey); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		log.Printf("Server running on http://localhost:%s", port)
		log.Println("WARNING: Running on HTTP. For production, set TLS_CERT and TLS_KEY environment variables.")
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}
