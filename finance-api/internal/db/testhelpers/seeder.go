package testhelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"time"
)

var (
	baseURL = "http://localhost:8282"
)

// Seeder contains a test database connection pool and HTTP server for API calls
type Seeder struct {
	Conn       *pgxpool.Pool
	HTTPClient *http.Client
}

// NewSeeder creates a new Seeder instance
func NewSeeder(ctx context.Context) *Seeder {
	conn, err := pgxpool.New(ctx, fmt.Sprintf("host=localhost port=5430 user=%s password=%s dbname=%s sslmode=disable", user, password, dbname))
	if err != nil {
		log.Fatal(err)
	}
	return &Seeder{
		Conn:       conn,
		HTTPClient: &http.Client{},
	}
}

func (s *Seeder) Exec(ctx context.Context, str string, i ...interface{}) (pgconn.CommandTag, error) {
	return s.Conn.Exec(ctx, str, i...)
}

func (s *Seeder) Query(ctx context.Context, str string, i ...interface{}) (pgx.Rows, error) {
	return s.Conn.Query(ctx, str, i...)
}

func (s *Seeder) QueryRow(ctx context.Context, str string, i ...interface{}) pgx.Row {
	return s.Conn.QueryRow(ctx, str, i...)
}

func (s *Seeder) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.Conn.BeginTx(ctx, pgx.TxOptions{})
}

func (s *Seeder) SeedData(data ...string) {
	ctx := context.Background()
	for _, d := range data {
		_, err := s.Exec(ctx, d)
		if err != nil {
			log.Fatal("Unable to seed data with db connection: " + err.Error())
		}
	}
}

func (s *Seeder) SendDataToAPI(ctx context.Context, method string, endpoint string, data interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", baseURL, endpoint)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

type DateHelper struct {
	date time.Time
}

func (s *Seeder) Today() DateHelper {
	date := DateHelper{date: time.Now()}
	return date
}

func (t DateHelper) Add(years int, months int, days int) DateHelper {
	t.date = t.date.AddDate(years, months, days)
	return t
}

func (t DateHelper) Sub(years int, months int, days int) DateHelper {
	t.date = t.date.AddDate(-years, -months, -days)
	return t
}

func (t DateHelper) String() string {
	return t.date.Format("2006-01-02")
}

func (t DateHelper) StringPtr() *string {
	s := t.date.Format("2006-01-02")
	return &s
}

func (t DateHelper) Date() time.Time {
	return t.date
}

func (t DateHelper) DatePtr() *time.Time {
	d := t.date
	return &d
}

func (t DateHelper) FinancialYear() string {
	if t.date.Month() >= time.April {
		return fmt.Sprintf("%d/%s", t.date.Year(), t.date.AddDate(1, 0, 0).Format("06"))
	}
	return fmt.Sprintf("%d/%s", t.date.Year()-1, t.date.Format("06"))
}
