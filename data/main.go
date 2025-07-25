package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/goinginblind/energy-sc-bot/data/internal/server"

	_ "github.com/lib/pq"
)

func main() {
	connStr := os.Getenv("DB_URI")
	if connStr == "" {
		log.Fatal("DB_URI env var required")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	srv := server.NewServer(db)
	log.Println("Starting REST API on :8080...")
	if err := http.ListenAndServe(":8080", srv.Router()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
