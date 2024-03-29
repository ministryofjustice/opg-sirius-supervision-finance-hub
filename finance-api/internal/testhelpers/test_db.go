package testhelpers

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"path/filepath"
	"runtime"
	"time"
)

const (
	dbname   = "test_db"
	user     = "test_user"
	password = "test_password"
)

var basePath string

// TestDatabase is a test utility containing a fully-migrated Postgres instance. To use this, run InitDb within a TestMain
// function and use the DbInstance to interact with the database as needed (e.g. to insert data prior to testing).
// Ensure to run TearDown at the end of the tests to clean up.
type TestDatabase struct {
	DbInstance *pgxpool.Pool
	DbAddress  string
	container  testcontainers.Container
}

func InitDb() *TestDatabase {
	ctx := context.Background()

	_, b, _, _ := runtime.Caller(0)
	testPath := filepath.Dir(b)
	basePath = filepath.Join(testPath, "../../..")

	container, err := postgres.RunContainer(
		ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase(dbname),
		postgres.WithUsername(user),
		postgres.WithPassword(password),
		postgres.WithInitScripts(basePath+"/migrations/000000_baseline.sql"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Fatal(err)
	}

	connString, err := container.ConnectionString(ctx)
	if err != nil {
		log.Fatal(err)
	}

	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatal(err)
	}

	err = migrateDb(connString)
	if err != nil {
		log.Fatal(err)
	}

	err = container.Snapshot(ctx, postgres.WithSnapshotName("test-snapshot"))
	if err != nil {
		log.Fatal(err)
	}

	return &TestDatabase{
		container:  container,
		DbInstance: db,
		DbAddress:  connString,
	}
}

func migrateDb(connString string) error {
	pathToMigrationFiles := basePath + "/migrations"

	m, err := migrate.New(fmt.Sprintf("file:%s", pathToMigrationFiles), fmt.Sprintf("%ssslmode=disable", connString))
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	log.Println("migration done")

	return nil
}

func (tdb *TestDatabase) TearDown() {
	tdb.DbInstance.Close()
	_ = tdb.container.Terminate(context.Background())
}
