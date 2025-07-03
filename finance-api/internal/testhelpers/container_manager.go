package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const (
	dbname   = "test_db"
	user     = "test_user"
	password = "test_password"
)

var basePath string

// ContainerManager is a test utility containing a fully-migrated Postgres instance. To use this, run Init within a TestMain
// function and use the DbInstance to interact with the database as needed (e.g. to insert data prior to testing).
// Ensure to run TearDown at the end of the tests to clean up.
type ContainerManager struct {
	Address   string
	Container *postgres.PostgresContainer
}

// Restore restores the DB to the snapshot backup and re-establishes the connection
func (db *ContainerManager) Restore(ctx context.Context) {
	err := db.Container.Restore(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Restore sometimes "completes" before indexes have been rebuilt, so we need to wait a bit
	time.Sleep(1 * time.Second)
}

func Init(ctx context.Context, searchPath string) *ContainerManager {
	_, b, _, _ := runtime.Caller(0)
	testPath := filepath.Dir(b)
	basePath = filepath.Join(testPath, "../../..")

	container, err := postgres.Run(
		ctx,
		"docker.io/postgres:13-alpine",
		postgres.WithDatabase(dbname),
		postgres.WithUsername(user),
		postgres.WithPassword(password),
		postgres.WithInitScripts(basePath+"/migrations/1_baseline.sql"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Fatal(err)
	}

	migrationConn, err := container.ConnectionString(ctx, "search_path=supervision_finance")
	if err != nil {
		log.Fatal(err)
	}

	err = migrateDb(ctx, migrationConn)
	if err != nil {
		log.Fatal(err)
	}

	err = container.Snapshot(ctx, postgres.WithSnapshotName("test-snapshot"))
	if err != nil {
		log.Fatal(err)
	}

	connString, err := container.ConnectionString(ctx, fmt.Sprintf("search_path=%s", searchPath))
	if err != nil {
		log.Fatal(err)
	}

	return &ContainerManager{
		Container: container,
		Address:   connString,
	}
}

func migrateDb(ctx context.Context, connString string) error {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		os.DirFS(basePath+"/migrations"),
		goose.WithExcludeNames([]string{"1_baseline.sql"}),
	)

	if err != nil {
		return err
	}

	if _, err = provider.Up(ctx); err != nil {
		return err
	}

	return nil
}

func (db *ContainerManager) TearDown(ctx context.Context) {
	_ = db.Container.Terminate(ctx)
}

func (db *ContainerManager) Seeder(ctx context.Context, t *testing.T) *Seeder {
	conn, err := pgxpool.New(ctx, db.Address)
	if err != nil {
		log.Fatal(err)
	}
	return &Seeder{
		t:    t,
		Conn: conn,
	}
}
