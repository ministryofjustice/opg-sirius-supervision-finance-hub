package service

import (
	"github.com/jackc/pgx/v5"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
)

type Service struct {
	Store *store.Queries
	DB    *pgx.Conn
}
