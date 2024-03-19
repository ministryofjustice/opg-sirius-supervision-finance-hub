package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/opg-sirius-finance-hub/finance-api/cmd/api"
	"github.com/opg-sirius-finance-hub/finance-api/internal/service"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Llongfile)

	dbUser := getEnv("POSTGRES_USER", "")
	dbPassword := getEnv("POSTGRES_PASSWORD", "")
	pgDb := getEnv("POSTGRES_DB", "")
	// Open a connection to the PostgreSQL database
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgresql://%s:%s@sirius-db:5432/%s?sslmode=disable", dbUser, dbPassword, pgDb))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	Store := store.New(conn)
	Service := service.Service{Store: Store}
	server := api.Server{Logger: logger, Service: &Service}

	server.SetupRoutes()

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
