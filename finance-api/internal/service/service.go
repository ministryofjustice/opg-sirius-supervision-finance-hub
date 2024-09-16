package service

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
)

type TX interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Service struct {
	store *store.Queries
	tx    TX
}

func NewService(conn *pgx.Conn) Service {
	return Service{
		store: store.New(conn),
		tx:    conn,
	}
}
