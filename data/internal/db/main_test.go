package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *Queries

func TestMain(m *testing.M) {
	conn, err := sql.Open("postgres", os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		log.Fatal("cannot connect to test database:", err)
	}

	testQueries = New(conn)

	os.Exit(m.Run())
}
