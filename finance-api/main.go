package main

import (
	"database/sql"
	"fmt"
	"github.com/opg-sirius-finance-hub/finance-api/internal"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbUser := getEnv("POSTGRES_USER", "")
	dbPassword := getEnv("POSTGRES_PASSWORD", "")
	pgDb := getEnv("POSTGRES_DB", "")
	// Open a connection to the PostgreSQL database
	db, err := sql.Open("postgres", fmt.Sprintf("postgresql://%s:%s@sirius-db:5432/%s?sslmode=disable", dbUser, dbPassword, pgDb))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Define a handler function to handle HTTP requests
	http.HandleFunc("/users/current", internal.GetCurrentUser(db))

	// Start the HTTP server on port 8080
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getEnv(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return def
}
