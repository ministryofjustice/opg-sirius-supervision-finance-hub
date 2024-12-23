package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/cmd/api"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/filestorage"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/reports"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/service"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	ctx := context.Background()
	logger := telemetry.NewLogger("opg-sirius-finance-api")

	err := run(ctx, logger)
	if err != nil {
		logger.Error("fatal startup error", slog.Any("err", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	exportTraces := os.Getenv("TRACING_ENABLED") == "1"
	shutdown, err := telemetry.StartTracerProvider(ctx, logger, exportTraces)
	defer shutdown()
	if err != nil {
		return err
	}

	writePool := setupDbPool(ctx, logger, "supervision_finance", false)
	defer writePool.Close()

	reportsClient := reports.NewClient(ctx, setupDbPool(ctx, logger, "supervision_finance,public", true))
	defer reportsClient.Close()

	eventClient := setupEventClient(ctx, logger)

	httpClient := http.DefaultClient

	notifyClient := notify.NewClient(httpClient)

	service := service.NewService(httpClient, writePool, reportsClient, eventClient, notifyClient)

	validator, err := validation.New()
	if err != nil {
		return err
	}
	fileStorage, err := filestorage.NewClient(ctx)
	if err != nil {
		return err
	}

	// TODO: Decide if Notify and reports goes here or in service
	server := api.NewServer(service, fileStorage, validator)

	s := &http.Server{
		Addr:    ":8080",
		Handler: server.SetupRoutes(logger),
	}

	go func() {
		if err := s.ListenAndServe(); err != nil {
			logger.Error("listen and server error", slog.Any("err", err.Error()))
			os.Exit(1)
		}
	}()
	logger.Info("Running at :8080")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	sig := <-c
	logger.Info("signal received: ", "sig", sig)

	tc, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return s.Shutdown(tc)
}

func setupDbPool(ctx context.Context, logger *slog.Logger, searchPath string, readOnly bool) *pgxpool.Pool {
	dbConn := os.Getenv("POSTGRES_CONN")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	pgDb := os.Getenv("POSTGRES_DB")

	connString := fmt.Sprintf("postgresql://%s:%s@%s/%s?search_path=%s", dbUser, url.QueryEscape(dbPassword), dbConn, pgDb, searchPath)
	if readOnly {
		connString += "&default_transaction_read_only=true"
	}

	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		logger.Error("Unable to create connection pool", "error", err)
		os.Exit(1)
	}
	return dbpool
}

func setupEventClient(ctx context.Context, logger *slog.Logger) *event.Client {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("failed to load aws config", slog.Any("err", err))
	}

	// set endpoint to "" outside dev to use default AWS resolver
	if endpointURL := os.Getenv("AWS_BASE_URL"); endpointURL != "" {
		cfg.BaseEndpoint = aws.String(endpointURL)
	}

	return event.NewClient(cfg, os.Getenv("EVENT_BUS_NAME"))
}
