// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: ledger_allocation.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createLedgerAllocation = `-- name: CreateLedgerAllocation :exec
WITH this_ledger as (
    SELECT id, datetime FROM ledger WHERE id = $1
)
INSERT INTO ledger_allocation (id, datetime, ledger_id, invoice_id, amount, status, notes)
SELECT nextval('ledger_allocation_id_seq'),
       this_ledger.datetime,
       $1,
       $2,
       $3,
       $4,
       $5
FROM this_ledger WHERE this_ledger.id = $1
`

type CreateLedgerAllocationParams struct {
	LedgerID  pgtype.Int4
	InvoiceID pgtype.Int4
	Amount    int32
	Status    string
	Notes     pgtype.Text
}

func (q *Queries) CreateLedgerAllocation(ctx context.Context, arg CreateLedgerAllocationParams) error {
	_, err := q.db.Exec(ctx, createLedgerAllocation,
		arg.LedgerID,
		arg.InvoiceID,
		arg.Amount,
		arg.Status,
		arg.Notes,
	)
	return err
}

const updateLedgerAllocationAdjustment = `-- name: UpdateLedgerAllocationAdjustment :exec
UPDATE ledger_allocation la
SET status = $1
FROM ledger l
WHERE l.id = $2
  AND l.id = la.ledger_id
  AND l.type IN ('CREDIT MEMO', 'CREDIT WRITE OFF', 'DEBIT MEMO')
`

type UpdateLedgerAllocationAdjustmentParams struct {
	Status string
	ID     int32
}

func (q *Queries) UpdateLedgerAllocationAdjustment(ctx context.Context, arg UpdateLedgerAllocationAdjustmentParams) error {
	_, err := q.db.Exec(ctx, updateLedgerAllocationAdjustment, arg.Status, arg.ID)
	return err
}
