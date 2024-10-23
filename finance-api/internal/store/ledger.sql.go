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
INSERT INTO ledger (id, datetime, finance_client_id, amount, notes, type, status, fee_reduction_id, created_at, created_by, reference, method)
SELECT nextval('ledger_id_seq'),
       now(),
       fc.id,
       $2,
       $3,
       $4,
       $5,
       $6,
       now(),
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
	CreatedBy      pgtype.Int4
}

func (q *Queries) CreateLedger(ctx context.Context, arg CreateLedgerParams) (int32, error) {
	row := q.db.QueryRow(ctx, createLedger,
		arg.ClientID,
		arg.Amount,
		arg.Notes,
		arg.Type,
		arg.Status,
		arg.FeeReductionID,
		arg.CreatedBy,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const createLedgerForCourtRef = `-- name: CreateLedgerForCourtRef :one
INSERT INTO ledger (id, datetime, finance_client_id, amount, notes, type, status, created_at, created_by, reference, method)
SELECT nextval('ledger_id_seq'),
       $2,
       fc.id,
       $3,
       $4,
       $5,
       $6,
       now(),
       $7,
       gen_random_uuid(),
       ''
FROM finance_client fc WHERE court_ref = $1
RETURNING id
`

type CreateLedgerForCourtRefParams struct {
	CourtRef  pgtype.Text
	Datetime  pgtype.Timestamp
	Amount    int32
	Notes     pgtype.Text
	Type      string
	Status    string
	CreatedBy pgtype.Int4
}

func (q *Queries) CreateLedgerForCourtRef(ctx context.Context, arg CreateLedgerForCourtRefParams) (int32, error) {
	row := q.db.QueryRow(ctx, createLedgerForCourtRef,
		arg.CourtRef,
		arg.Datetime,
		arg.Amount,
		arg.Notes,
		arg.Type,
		arg.Status,
		arg.CreatedBy,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const getLedgerForPayment = `-- name: GetLedgerForPayment :one
SELECT l.id
FROM ledger l
LEFT JOIN finance_client fc ON fc.id = l.finance_client_id
WHERE l.amount = $1 AND l.status = 'CONFIRMED' AND l.datetime = $2 AND l.type = $3 AND fc.court_ref = $4
LIMIT 1
`

type GetLedgerForPaymentParams struct {
	Amount   int32
	Datetime pgtype.Timestamp
	Type     string
	CourtRef pgtype.Text
}

func (q *Queries) GetLedgerForPayment(ctx context.Context, arg GetLedgerForPaymentParams) (int32, error) {
	row := q.db.QueryRow(ctx, getLedgerForPayment,
		arg.Amount,
		arg.Datetime,
		arg.Type,
		arg.CourtRef,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}
