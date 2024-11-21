//go:build seed && !release

package seed

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/service"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	ctx := context.Background()

	publicDb, err := setupDbPool(ctx, "public,supervision_finance")
	if err != nil {
		log.Fatalf("failed to setup public db pool: %v", err)
	}
	defer publicDb.Close()

	_ = publicSchemaClient{publicDb}

	financeDb, err := setupDbPool(ctx, "supervision_finance")
	if err != nil {
		log.Fatalf("failed to setup supervision_finance db pool: %v", err)
	}
	defer financeDb.Close()

	_ = service.NewService(http.DefaultClient, financeDb, &dispatchStub{}, &fileStorageStub{})

	//// Seed data
	//if err := seedData(ctx, svc); err != nil {
	//	logger.Error("failed to seed data", slog.Any("err", err))
	//	os.Exit(1)
	//}

	log.Println("seed data complete")
}

func setupDbPool(ctx context.Context, searchPath string) (*pgxpool.Pool, error) {
	dbConn := os.Getenv("POSTGRES_CONN")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	pgDb := os.Getenv("POSTGRES_DB")

	return pgxpool.New(ctx, fmt.Sprintf("postgresql://%s:%s@%s/%s?search_path=%s", dbUser, url.QueryEscape(dbPassword), dbConn, pgDb, searchPath))
}
