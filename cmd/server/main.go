package main

import (
	"log"
	"net/http"
	"os"

	"ledger-go-system/internal/db"
	"ledger-go-system/internal/handler"
	"ledger-go-system/internal/middleware"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	conn, err := db.New(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	h := handler.NewLedgerHandler(conn)

	mux := http.NewServeMux()
	mux.Handle("POST /ledger", middleware.Role("admin", http.HandlerFunc(h.Create)))
	mux.Handle("GET /ledger", middleware.RoleAny(http.HandlerFunc(h.List)))
	mux.Handle("GET /ledger/", middleware.RoleAny(http.HandlerFunc(h.GetByID)))

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
