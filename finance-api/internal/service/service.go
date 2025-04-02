package service

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"io"
	"log/slog"
	"net/http"
)

type TX interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Dispatch interface {
	CreditOnAccount(ctx context.Context, event event.CreditOnAccount) error
	PaymentMethodChanged(ctx context.Context, event event.PaymentMethod) error
}

type FileStorage interface {
	GetFile(ctx context.Context, bucketName string, fileName string) (io.ReadCloser, error)
	GetFileWithVersion(ctx context.Context, bucketName string, fileName string, versionID string) (io.ReadCloser, error)
	PutFile(ctx context.Context, bucketName string, fileName string, file io.Reader) (*string, error)
}

type NotifyClient interface {
	Send(ctx context.Context, payload notify.Payload) error
}

type Env struct {
	AsyncBucket string
}

type Service struct {
	store       *store.Queries
	dispatch    Dispatch
	fileStorage FileStorage
	notify      NotifyClient
	tx          TX
	env         *Env
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewService(conn *pgxpool.Pool, dispatch Dispatch, fileStorage FileStorage, notify NotifyClient, env *Env) *Service {
	return &Service{
		store:       store.New(conn),
		dispatch:    dispatch,
		fileStorage: fileStorage,
		notify:      notify,
		tx:          conn,
		env:         env,
	}
}

func (s *Service) BeginStoreTx(ctx context.Context) (*store.Tx, error) {
	tx, err := s.tx.Begin(ctx)
	logger := s.Logger(ctx)
	if err != nil {
		logger.Error("Could not begin a transaction", slog.String("err", err.Error()))
		return nil, err
	}

	return store.NewTx(tx), nil
}

func (s *Service) Logger(ctx context.Context) *slog.Logger {
	return telemetry.LoggerFromContext(ctx)
}

func (*Service) WithCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	cancelCtx, cancelTx := context.WithCancel(ctx)
	return ctx.(auth.Context).WithContext(cancelCtx), cancelTx
}
