package service

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opg-sirius-finance-hub/finance-api/internal/event"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
)

type TX interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Service struct {
	store    *store.Queries
	dispatch *event.Client
	tx       TX
}

func NewService(conn *pgxpool.Pool, eventClient *event.Client) Service {
	return Service{
		store:    store.New(conn),
		dispatch: eventClient,
		tx:       conn,
	}
}
