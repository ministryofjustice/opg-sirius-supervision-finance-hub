package testhelpers

import (
	"context"
	"database/sql"
	"github.com/jackc/pgx/v5"
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
	DbConn     *pgx.Conn
}

func (db *TestDatabase) SeedData(data ...string) {
	ctx := context.Background()
	for _, d := range data {
		_, err := db.DbInstance.Exec(ctx, d)
		if err != nil {
			log.Fatal("Unable to seed data with db connection: " + err.Error())
		}
	}
}

func InitDb() *TestDatabase {
	ctx := context.Background()

	_, b, _, _ := runtime.Caller(0)
	testPath := filepath.Dir(b)
	basePath = filepath.Join(testPath, "../../..")

	container, err := postgres.RunContainer(
		ctx,
		testcontainers.WithImage("docker.io/postgres:13-alpine"),
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

	connString, err := container.ConnectionString(ctx, "search_path=supervision_finance")
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

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		log.Fatal(err)
	}

	return &TestDatabase{
		container:  container,
		DbInstance: db,
		DbAddress:  connString,
		DbConn:     conn,
	}
}

func migrateDb(connString string) error {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return err
	}
	defer db.Close()

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		os.DirFS(basePath+"/migrations"),
		goose.WithExcludeNames([]string{"1_baseline.sql"}),
	)

	if err != nil {
		return err
	}

	if _, err = provider.Up(context.Background()); err != nil {
		return err
	}

	return nil
}

func (db *TestDatabase) TearDown() {
	db.DbInstance.Close()
	_ = db.container.Terminate(context.Background())
}
