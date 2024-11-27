package service

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
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
}

type Service struct {
	http        HTTPClient
	store       *store.Queries
	dispatch    Dispatch
	filestorage FileStorage
	tx          TX
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewService(httpClient HTTPClient, conn *pgxpool.Pool, dispatch Dispatch, filestorage FileStorage) Service {
	return Service{
		http:        httpClient,
		store:       store.New(conn),
		dispatch:    dispatch,
		filestorage: filestorage,
		tx:          conn,
	}
}
