package service

import (
	"github.com/jackc/pgx/v5"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
)

type BadRequest struct {
	Reason string
}

func (b BadRequest) Error() string {
	return b.Reason
}

type Service struct {
	Store *store.Queries
	DB    *pgx.Conn
}
