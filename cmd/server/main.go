package main

import (
	"log"
	"net/http"
	"os"

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

	conn, err := db.New(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	authManager := auth.NewAuthManager(jwtSecret)
	ledgerHandler := handler.NewLedgerHandler(conn)
	authHandler := handler.NewAuthHandler(authManager)

	mux := http.NewServeMux()

	// Public endpoints (no auth required)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// Protected endpoints (require JWT)
	// Admin only: POST /ledger
	mux.Handle("POST /ledger", middleware.RequireRole("admin", authManager, http.HandlerFunc(ledgerHandler.Create)))

	// Both admin and viewer: GET /ledger, GET /ledger/{id}
	mux.Handle("GET /ledger", middleware.AllowRoles([]string{"admin", "viewer"}, authManager, http.HandlerFunc(ledgerHandler.List)))
	mux.Handle("GET /ledger/", middleware.AllowRoles([]string{"admin", "viewer"}, authManager, http.HandlerFunc(ledgerHandler.GetByID)))

	log.Printf("Server running on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
