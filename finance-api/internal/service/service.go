package service

import (
	"github.com/jackc/pgx/v5"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/finance-api/internal/validation"
)

type Service struct {
	Store  *store.Queries
	VError *validation.Validate
	DB     *pgx.Conn
}
