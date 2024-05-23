// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: ledger.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createLedgerForFeeReduction = `-- name: CreateLedgerForFeeReduction :one
insert into ledger (id, reference, datetime, method, amount, notes, type, status, finance_client_id,
                                        parent_id, fee_reduction_id, confirmeddate, bankdate, batchnumber, bankaccount,
                                        line, source,
                                        createddate, createdby_id)
VALUES (nextval('ledger_id_seq'::regclass), gen_random_uuid(), now(), $1, $2, $3, $4, 'Status', $5, null, $6, null,
        null, null, null, null, null, now(), $7) returning id, reference, datetime, method, amount, notes, type, status, finance_client_id, parent_id, fee_reduction_id, confirmeddate, bankdate, batchnumber, bankaccount, source, line, createddate, createdby_id
`

type CreateLedgerForFeeReductionParams struct {
	Method          string
	Amount          int32
	Notes           pgtype.Text
	Type            string
	FinanceClientID pgtype.Int4
	FeeReductionID  pgtype.Int4
	CreatedbyID     pgtype.Int4
}

func (q *Queries) CreateLedgerForFeeReduction(ctx context.Context, arg CreateLedgerForFeeReductionParams) (Ledger, error) {
	row := q.db.QueryRow(ctx, createLedgerForFeeReduction,
		arg.Method,
		arg.Amount,
		arg.Notes,
		arg.Type,
		arg.FinanceClientID,
		arg.FeeReductionID,
		arg.CreatedbyID,
	)
	var i Ledger
	err := row.Scan(
		&i.ID,
		&i.Reference,
		&i.Datetime,
		&i.Method,
		&i.Amount,
		&i.Notes,
		&i.Type,
		&i.Status,
		&i.FinanceClientID,
		&i.ParentID,
		&i.FeeReductionID,
		&i.Confirmeddate,
		&i.Bankdate,
		&i.Batchnumber,
		&i.Bankaccount,
		&i.Source,
		&i.Line,
		&i.Createddate,
		&i.CreatedbyID,
	)
	return i, err
}

const updateLedgerAdjustment = `-- name: UpdateLedgerAdjustment :exec
UPDATE ledger l
SET status = 'APPROVED'
WHERE l.id = $1
`

func (q *Queries) UpdateLedgerAdjustment(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, updateLedgerAdjustment, id)
	return err
}
