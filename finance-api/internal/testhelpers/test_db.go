package testhelpers

import (
	"context"
	"database/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	Address   string
	Container *postgres.PostgresContainer
}

// Restore restores the DB to the snapshot backup and re-establishes the connection
func (db *TestDatabase) Restore() {
	err := db.Container.Restore(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func InitDb() *TestDatabase {
	ctx := context.Background()

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

	connString, err := container.ConnectionString(ctx, "search_path=supervision_finance")
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
		Container: container,
		Address:   connString,
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
	_ = db.Container.Terminate(context.Background())
}

func (db *TestDatabase) GetConn() TestConn {
	conn, err := pgx.Connect(context.Background(), db.Address)
	if err != nil {
		log.Fatal(err)
	}
	return TestConn{conn}
}

type TestConn struct {
	Conn *pgx.Conn
}

func (c TestConn) Exec(ctx context.Context, s string, i ...interface{}) (pgconn.CommandTag, error) {
	return c.Conn.Exec(ctx, s, i...)
}

func (c TestConn) Query(ctx context.Context, s string, i ...interface{}) (pgx.Rows, error) {
	return c.Conn.Query(ctx, s, i...)
}

func (c TestConn) QueryRow(ctx context.Context, s string, i ...interface{}) pgx.Row {
	return c.Conn.QueryRow(ctx, s, i...)
}

func (c TestConn) Prepare(ctx context.Context, name string, sql string) (sd *pgconn.StatementDescription, err error) {
	return c.Conn.Prepare(ctx, name, sql)
}

func (c TestConn) Begin(ctx context.Context) (pgx.Tx, error) {
	return c.Conn.BeginTx(ctx, pgx.TxOptions{})
}

func (c TestConn) SeedData(data ...string) {
	ctx := context.Background()
	for _, d := range data {
		_, err := c.Exec(ctx, d)
		if err != nil {
			log.Fatal("Unable to seed data with db connection: " + err.Error())
		}
	}
}
