// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: ledger.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createLedger = `-- name: CreateLedger :one
INSERT INTO ledger (id, datetime, finance_client_id, amount, notes, type, status, fee_reduction_id, createdby_id, reference, method)
SELECT nextval('ledger_id_seq'),
       now(),
       fc.id,
       $2,
       $3,
       $4,
       $5,
       $6,
       $7,
       gen_random_uuid(),
       ''
FROM finance_client fc WHERE client_id = $1
RETURNING id
`

type CreateLedgerParams struct {
	ClientID       int32
	Amount         int32
	Notes          pgtype.Text
	Type           string
	Status         string
	FeeReductionID pgtype.Int4
	CreatedbyID    pgtype.Int4
}

func (q *Queries) CreateLedger(ctx context.Context, arg CreateLedgerParams) (int32, error) {
	row := q.db.QueryRow(ctx, createLedger,
		arg.ClientID,
		arg.Amount,
		arg.Notes,
		arg.Type,
		arg.Status,
		arg.FeeReductionID,
		arg.CreatedbyID,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const updateLedgerAdjustment = `-- name: UpdateLedgerAdjustment :exec
UPDATE ledger l
SET status = $1
WHERE l.id = $2
`

type UpdateLedgerAdjustmentParams struct {
	Status string
	ID     int32
}

func (q *Queries) UpdateLedgerAdjustment(ctx context.Context, arg UpdateLedgerAdjustmentParams) error {
	_, err := q.db.Exec(ctx, updateLedgerAdjustment, arg.Status, arg.ID)
	return err
}
