package reports

import (
	"context"
	"encoding/csv"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/db"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"io"
	"os"
	"path/filepath"
	"time"
)

type dbClient interface {
	Run(ctx context.Context, query db.ReportQuery) ([][]string, error)
	Close()
}

type fileStorageClient interface {
	PutFile(ctx context.Context, bucketName string, fileName string, file io.Reader) (*string, error)
}

type notifyClient interface {
	Send(ctx context.Context, payload notify.Payload) error
}

type Envs struct {
	ReportsBucket   string
	FinanceAdminURL string
	GoLiveDate      time.Time
}

type Client struct {
	db          dbClient
	fileStorage fileStorageClient
	notify      notifyClient
	envs        *Envs
}

func (c *Client) Close() {
	c.db.Close()
}

func NewClient(dbPool *pgxpool.Pool, fileStorage fileStorageClient, notify notifyClient, envs *Envs) *Client {
	return &Client{
		db:          db.NewClient(dbPool),
		fileStorage: fileStorage,
		notify:      notify,
		envs:        envs,
	}
}

func (c *Client) generate(ctx context.Context, filename string, query db.ReportQuery) (*os.File, error) {
	rows, err := c.db.Run(ctx, query)
	if err != nil {
		return nil, err
	}

	return createCsv(filename, rows)
}

func createCsv(filename string, items [][]string) (*os.File, error) {
	file, err := os.Create(filepath.Clean(filename))
	if err != nil {
		return nil, err
	}

	defer file.Close()

	writer := csv.NewWriter(file)

	for _, item := range items {
		err = writer.Write(item)
		if err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if writer.Error() != nil {
		return nil, writer.Error()
	}

	return os.Open(filename)
}
