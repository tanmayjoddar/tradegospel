package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/prisma/prisma-client-go/runtime/httpclient"

	"ledger-go-system/internal/db"
	"ledger-go-system/internal/handler"
	"ledger-go-system/internal/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	client, err := db.New()
	if err != nil {
		log.Fatalf("Failed to initialize Prisma client: %v", err)
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Error disconnecting Prisma client: %v", err)
		}
	}()

	h := handler.NewLedgerHandler(client)

	mux := http.NewServeMux()
	mux.Handle("POST /ledger", middleware.Role("admin", http.HandlerFunc(h.Create)))
	mux.Handle("GET /ledger", middleware.RoleAny(http.HandlerFunc(h.List)))
	mux.Handle("GET /ledger/", middleware.RoleAny(http.HandlerFunc(h.GetByID)))

	log.Printf("Server running on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
