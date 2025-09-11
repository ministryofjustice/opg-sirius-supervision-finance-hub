package store

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Tx is a wrapper around pgx.Tx that adds the transaction to the Queries, providing a method to commit
type Tx struct {
	*Queries
	tx pgx.Tx
}

func NewTx(tx pgx.Tx) *Tx {
	return &Tx{
		Queries: New(tx),
		tx:      tx,
	}
}

func (s *Tx) Commit(ctx context.Context) error {
	return s.tx.Commit(ctx)
}

func (s *Tx) Rollback(ctx context.Context) {
	_ = s.tx.Rollback(ctx)
}
