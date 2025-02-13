package main

import (
	"context"
	"errors"
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
	"strconv"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
)

type Envs struct {
	webDir             string
	siriusPublicURL    string
	awsRegion          string
	iamRole            string
	s3Endpoint         string
	s3EncryptionKey    string
	notifyKey          string
	asyncBucket        string
	goLiveDate         string
	reportsBucket      string
	financeAdminPrefix string
	dbConn             string
	dbUser             string
	dbPassword         string
	dbName             string
	awsBaseUrl         string
	eventBusName       string
	port               string
	jwtSecret          string
	jwtExpiry          int
}

func parseEnvs() (*Envs, error) {
	envs := map[string]string{
		"AWS_REGION":            os.Getenv("AWS_REGION"),
		"AWS_IAM_ROLE":          os.Getenv("AWS_IAM_ROLE"),
		"AWS_S3_ENDPOINT":       os.Getenv("AWS_S3_ENDPOINT"),
		"S3_ENCRYPTION_KEY":     os.Getenv("S3_ENCRYPTION_KEY"),
		"JWT_SECRET":            os.Getenv("JWT_SECRET"),
		"JWT_EXPIRY":            os.Getenv("JWT_EXPIRY"),
		"OPG_NOTIFY_API_KEY":    os.Getenv("OPG_NOTIFY_API_KEY"),
		"ASYNC_S3_BUCKET":       os.Getenv("ASYNC_S3_BUCKET"),
		"FINANCE_HUB_LIVE_DATE": os.Getenv("FINANCE_HUB_LIVE_DATE"),
		"REPORTS_S3_BUCKET":     os.Getenv("REPORTS_S3_BUCKET"),
		"SIRIUS_PUBLIC_URL":     os.Getenv("SIRIUS_PUBLIC_URL"),
		"FINANCE_ADMIN_PREFIX":  os.Getenv("FINANCE_ADMIN_PREFIX"),
		"POSTGRES_CONN":         os.Getenv("POSTGRES_CONN"),
		"POSTGRES_USER":         os.Getenv("POSTGRES_USER"),
		"POSTGRES_PASSWORD":     os.Getenv("POSTGRES_PASSWORD"),
		"POSTGRES_DB":           os.Getenv("POSTGRES_DB"),
		"EVENT_BUS_NAME":        os.Getenv("EVENT_BUS_NAME"),
	}

	var missing []error
	for k, v := range envs {
		if v == "" {
			missing = append(missing, errors.New("missing environment variable: "+k))
		}
	}

	jwtExpiry, err := strconv.Atoi(envs["JWT_EXPIRY"])
	if err != nil {
		missing = append(missing, errors.New("invalid JWT_EXPIRY"))
	}

	if len(missing) > 0 {
		return nil, errors.Join(missing...)
	}

	return &Envs{
		awsRegion:          envs["AWS_REGION"],
		iamRole:            envs["AWS_IAM_ROLE"],
		s3Endpoint:         envs["AWS_S3_ENDPOINT"],
		s3EncryptionKey:    envs["S3_ENCRYPTION_KEY"],
		jwtSecret:          envs["JWT_SECRET"],
		notifyKey:          envs["OPG_NOTIFY_API_KEY"],
		asyncBucket:        envs["ASYNC_S3_BUCKET"],
		goLiveDate:         envs["FINANCE_HUB_LIVE_DATE"],
		reportsBucket:      envs["REPORTS_S3_BUCKET"],
		siriusPublicURL:    envs["SIRIUS_PUBLIC_URL"],
		financeAdminPrefix: envs["FINANCE_ADMIN_PREFIX"],
		dbConn:             envs["POSTGRES_CONN"],
		dbUser:             envs["POSTGRES_USER"],
		dbPassword:         envs["POSTGRES_PASSWORD"],
		dbName:             envs["POSTGRES_DB"],
		awsBaseUrl:         os.Getenv("AWS_BASE_URL"), // can be empty
		eventBusName:       envs["EVENT_BUS_NAME"],
		jwtExpiry:          jwtExpiry,
		webDir:             "web",
		port:               "8888",
	}, nil
}

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

	envs, err := parseEnvs()
	if err != nil {
		return err
	}

	dbPool := setupDbPool(ctx, logger, "supervision_finance", envs, false)
	defer dbPool.Close()

	eventClient := setupEventClient(ctx, logger, envs)
	fileStorageClient, err := filestorage.NewClient(
		ctx,
		envs.awsRegion,
		envs.iamRole,
		envs.s3Endpoint,
		envs.s3EncryptionKey,
	)

	if err != nil {
		return err
	}

	notifyClient := notify.NewClient(envs.notifyKey)

	Service := service.NewService(
		dbPool,
		eventClient,
		fileStorageClient,
		notifyClient,
		&service.Env{
			AsyncBucket: envs.asyncBucket,
		},
	)

	validator, err := validation.New()
	if err != nil {
		return err
	}

	goLiveDate, _ := time.Parse("2006-01-02", envs.goLiveDate)

	reportsClient := reports.NewClient(
		setupDbPool(ctx, logger, "supervision_finance,public", nil, true),
		fileStorageClient,
		notifyClient,
		&reports.Envs{
			ReportsBucket:   envs.reportsBucket,
			FinanceAdminURL: fmt.Sprintf("%s%s", envs.siriusPublicURL, envs.financeAdminPrefix),
			GoLiveDate:      goLiveDate,
		},
	)
	defer reportsClient.Close()

	server := api.NewServer(Service, reportsClient, fileStorageClient, validator, &api.Envs{
		ReportsBucket: envs.reportsBucket,
		GoLiveDate:    goLiveDate,
	})

	s := &http.Server{
		Addr:    ":" + envs.port,
		Handler: server.SetupRoutes(logger),
	}

	go func() {
		if err := s.ListenAndServe(); err != nil {
			logger.Error("listen and server error", slog.Any("err", err.Error()))
			os.Exit(1)
		}
	}()
	logger.Info("Running at :" + envs.port)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	sig := <-c
	logger.Info("signal received: ", "sig", sig)

	tc, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return s.Shutdown(tc)
}

func setupDbPool(ctx context.Context, logger *slog.Logger, searchPath string, envs *Envs, readOnly bool) *pgxpool.Pool {
	connString := fmt.Sprintf("postgresql://%s:%s@%s/%s?search_path=%s", envs.dbUser, url.QueryEscape(envs.dbPassword), envs.dbConn, envs.dbName, searchPath)
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

func setupEventClient(ctx context.Context, logger *slog.Logger, envs *Envs) *event.Client {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("failed to load aws config", slog.Any("err", err))
	}

	// set endpoint to "" outside dev to use default AWS resolver
	if envs.awsBaseUrl != "" {
		cfg.BaseEndpoint = aws.String(envs.awsBaseUrl)
	}

	return event.NewClient(cfg, envs.eventBusName)
}
