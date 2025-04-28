package reports

import (
	"bytes"
	"context"
	"encoding/csv"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/db"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"io"
	"time"
)

type dbClient interface {
	Run(ctx context.Context, query db.ReportQuery) ([][]string, error)
	CopyStream(ctx context.Context, query db.ReportQuery) (io.ReadCloser, error)
	Close()
}

type fileStorageClient interface {
	StreamFile(ctx context.Context, bucketName string, fileName string, stream io.ReadCloser) (*string, error)
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

func (c *Client) stream(ctx context.Context, query db.ReportQuery) (io.ReadCloser, error) {
	return c.db.CopyStream(ctx, query)
}

func (c *Client) generate(ctx context.Context, query db.ReportQuery) (*bytes.Buffer, error) {
	rows, err := c.db.Run(ctx, query)
	if err != nil {
		return nil, err
	}

	return createCsv(rows)
}

func createCsv(items [][]string) (*bytes.Buffer, error) {
	var buffer bytes.Buffer

	writer := csv.NewWriter(&buffer)

	for _, item := range items {
		err := writer.Write(item)
		if err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if writer.Error() != nil {
		return nil, writer.Error()
	}

	return &buffer, nil
}
