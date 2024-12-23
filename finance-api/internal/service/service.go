package service

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/reports"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"net/http"
)

type TX interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Dispatch interface {
	CreditOnAccount(ctx context.Context, event event.CreditOnAccount) error
	FinanceAdminUploadProcessed(ctx context.Context, event event.FinanceAdminUploadProcessed) error
}

type Notify interface {
	Send(ctx context.Context, message notify.Payload) error
}

type Service struct {
	http     HTTPClient
	store    *store.Queries
	reports  *reports.Client
	dispatch Dispatch
	notify   Notify
	tx       TX
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewService(httpClient HTTPClient, writePool *pgxpool.Pool, reports *reports.Client, dispatch Dispatch, notify Notify) *Service {
	return &Service{
		http:     httpClient,
		store:    store.New(writePool),
		reports:  reports,
		dispatch: dispatch,
		notify:   notify,
		tx:       writePool,
	}
}
