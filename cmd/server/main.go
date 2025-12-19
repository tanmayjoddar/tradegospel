package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

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

	conn, err := db.New(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	h := handler.NewLedgerHandler(conn)

	mux := http.NewServeMux()
	mux.Handle("POST /ledger", middleware.Role("admin", http.HandlerFunc(h.Create)))
	mux.Handle("GET /ledger", middleware.RoleAny(http.HandlerFunc(h.List)))
	mux.Handle("GET /ledger/", middleware.RoleAny(http.HandlerFunc(h.GetByID)))

	log.Printf("Server running on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
