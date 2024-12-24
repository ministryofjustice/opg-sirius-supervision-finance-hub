package service

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/reports"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"io"
	"net/http"
)

type TX interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Dispatch interface {
	CreditOnAccount(ctx context.Context, event event.CreditOnAccount) error
	FinanceAdminUploadProcessed(ctx context.Context, event event.FinanceAdminUploadProcessed) error
}

type FileStorage interface {
	GetFile(ctx context.Context, bucketName string, fileName string) (io.ReadCloser, error)
	PutFile(ctx context.Context, bucketName string, fileName string, file io.Reader) (*string, error)
}

type Service struct {
	store       *store.Queries
	dispatch    Dispatch
	fileStorage FileStorage
	reports     *reports.Client
	notify      *notify.Client
	tx          TX
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewService(conn *pgxpool.Pool, dispatch Dispatch, fileStorage FileStorage, reports *reports.Client, notify *notify.Client) Service {
	return Service{
		store:       store.New(conn),
		dispatch:    dispatch,
		fileStorage: fileStorage,
		reports:     reports,
		notify:      notify,
		tx:          conn,
	}
}
