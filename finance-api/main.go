package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/ministryofjustice/opg-go-common/env"
	"github.com/opg-sirius-finance-hub/auth"
	"github.com/opg-sirius-finance-hub/finance-api/cmd/api"
	"github.com/opg-sirius-finance-hub/finance-api/internal/service"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"go.opentelemetry.io/contrib/detectors/aws/ecs"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

func initTracerProvider(ctx context.Context, logger *zap.SugaredLogger) func() {
	resource, err := ecs.NewResourceDetector().Detect(ctx)
	if err != nil {
		logger.Fatal(err)
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("0.0.0.0:4317"),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		logger.Fatal(err)
	}

	idg := xray.NewIDGenerator()
	tp := trace.NewTracerProvider(
		trace.WithResource(resource),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(traceExporter),
		trace.WithIDGenerator(idg),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Fatal(err)
		}
	}
}

func main() {
	logger := zap.Must(zap.NewProduction(zap.Fields(zap.String("service_name", "opg-sirius-finance-api")))).Sugar()

	defer func() { _ = logger.Sync() }()

	if env.Get("TRACING_ENABLED", "0") == "1" {
		shutdown := initTracerProvider(context.Background(), logger)
		defer shutdown()
	}

	dbConn := getEnv("POSTGRES_CONN", "")
	dbUser := getEnv("POSTGRES_USER", "")
	dbPassword := getEnv("POSTGRES_PASSWORD", "")
	pgDb := getEnv("POSTGRES_DB", "")
	// Open a connection to the PostgreSQL database
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgresql://%s:%s@%s/%s?search_path=supervision_finance", dbUser, dbPassword, dbConn, pgDb))
	if err != nil {
		logger.Fatal(err)
	}
	defer conn.Close(ctx)

	jwtEnabled := getEnv("TOGGLE_JWT_ENABLED", "0") == "1"
	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY", "1"))
	jwtSecret := getEnv("JWT_SECRET", "mysupersecrettestkeythatis128bits")
	jwtConfig := auth.JwtConfig{Enabled: jwtEnabled, Secret: jwtSecret, Expiry: jwtExpiry}

	Store := store.New(conn)
	Service := service.Service{Store: Store}
	server := api.Server{Logger: logger, Service: &Service, JwtConfig: jwtConfig}

	server.SetupRoutes()

	// Start the HTTP server on port 8080
	logger.Infow("Server listening on :8080")
	logger.Fatal(http.ListenAndServe(":8080", nil))
}

func getEnv(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return def
}
