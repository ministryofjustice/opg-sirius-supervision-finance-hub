package service

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"io"
	"net/http"
)

type TX interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Dispatch interface {
	CreditOnAccount(ctx context.Context, event event.CreditOnAccount) error
}

type FileStorage interface {
	GetFile(ctx context.Context, bucketName string, fileName string) (io.ReadCloser, error)
	PutFile(ctx context.Context, bucketName string, fileName string, file io.Reader) (*string, error)
}

type NotifyClient interface {
	Send(ctx context.Context, payload notify.Payload) error
}

type Service struct {
	store       *store.Queries
	dispatch    Dispatch
	fileStorage FileStorage
	notify      NotifyClient
	tx          TX
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewService(conn *pgxpool.Pool, dispatch Dispatch, fileStorage FileStorage, notify NotifyClient) *Service {
	return &Service{
		store:       store.New(conn),
		dispatch:    dispatch,
		fileStorage: fileStorage,
		notify:      notify,
		tx:          conn,
	}
}

func (s *Service) BeginStoreTx(ctx context.Context) (*store.Tx, error) {
	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return store.NewTx(tx), nil
}
