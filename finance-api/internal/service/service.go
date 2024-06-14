package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"strings"
)

type BadRequest struct {
	Reason string
}

func (b BadRequest) Error() string {
	return b.Reason
}

type BadRequests struct {
	Reasons []string `json:"reasons"`
}

func (b BadRequests) Error() string {
	return fmt.Sprintf("bad requests: %s", strings.Join(b.Reasons, ", "))
}

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
