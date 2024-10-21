package service

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opg-sirius-finance-hub/finance-api/internal/event"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"io"
)

type TX interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Dispatch interface {
	CreditOnAccount(ctx context.Context, event event.CreditOnAccount) error
}

type FileStorage interface {
	GetFile(ctx context.Context, bucketName string, fileName string) (io.ReadCloser, error)
}

type Service struct {
	store       *store.Queries
	dispatch    Dispatch
	filestorage FileStorage
	tx          TX
}

func NewService(conn *pgxpool.Pool, dispatch Dispatch, filestorage FileStorage) Service {
	return Service{
		store:       store.New(conn),
		dispatch:    dispatch,
		filestorage: filestorage,
		tx:          conn,
	}
}
