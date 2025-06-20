// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: ledger_allocation.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createLedgerAllocation = `-- name: CreateLedgerAllocation :exec
WITH this_ledger AS (SELECT id, datetime
                     FROM ledger l
                     WHERE l.id = $4::INT),
     allocation AS (INSERT INTO ledger_allocation (id, datetime, ledger_id, invoice_id, amount, status, notes)
         SELECT NEXTVAL('ledger_allocation_id_seq'),
                this_ledger.datetime,
                $4::INT,
                $3::INT,
                $2::INT,
                $1::TEXT,
                $5
         FROM this_ledger
         WHERE this_ledger.id = $4::INT)
UPDATE invoice i
SET cacheddebtamount = CASE
                           WHEN $1::TEXT = 'UNAPPLIED' THEN cacheddebtamount
                           ELSE COALESCE(cacheddebtamount, i.amount) - $2::INT END
WHERE $3::INT IS NOT NULL AND i.id = $3::INT
`

type CreateLedgerAllocationParams struct {
	Status    string
	Amount    int32
	InvoiceID pgtype.Int4
	LedgerID  int32
	Notes     pgtype.Text
}

func (q *Queries) CreateLedgerAllocation(ctx context.Context, arg CreateLedgerAllocationParams) error {
	_, err := q.db.Exec(ctx, createLedgerAllocation,
		arg.Status,
		arg.Amount,
		arg.InvoiceID,
		arg.LedgerID,
		arg.Notes,
	)
	return err
}
