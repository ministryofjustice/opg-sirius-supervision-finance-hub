// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: invoice_adjustments.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createLedgerForAdjustment = `-- name: CreateLedgerForAdjustment :one
WITH created AS (
    INSERT INTO ledger (id, datetime, finance_client_id, amount, notes, type, status, fee_reduction_id, created_at,
                        created_by, reference, method)
        SELECT NEXTVAL('ledger_id_seq'),
               NOW(),
               fc.id,
               $2,
               $3,
               $4,
               $5,
               $6,
               NOW(),
               $7,
               gen_random_uuid(),
               ''
        FROM finance_client fc
        WHERE client_id = $1
        RETURNING id)
UPDATE invoice_adjustment ia
SET ledger_id = created.id
FROM created
WHERE ia.id = $8
RETURNING created.id
`

type CreateLedgerForAdjustmentParams struct {
	ClientID       int32
	Amount         int32
	Notes          pgtype.Text
	Type           string
	Status         string
	FeeReductionID pgtype.Int4
	CreatedBy      pgtype.Int4
	ID             int32
}

func (q *Queries) CreateLedgerForAdjustment(ctx context.Context, arg CreateLedgerForAdjustmentParams) (int32, error) {
	row := q.db.QueryRow(ctx, createLedgerForAdjustment,
		arg.ClientID,
		arg.Amount,
		arg.Notes,
		arg.Type,
		arg.Status,
		arg.FeeReductionID,
		arg.CreatedBy,
		arg.ID,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const createPendingInvoiceAdjustment = `-- name: CreatePendingInvoiceAdjustment :one
INSERT INTO invoice_adjustment (id, finance_client_id, invoice_id, raised_date, adjustment_type, amount, notes, status,
                                created_at, created_by)
SELECT NEXTVAL('invoice_adjustment_id_seq'),
       fc.id,
       $2,
       NOW(),
       $3,
       $4,
       $5,
       'PENDING',
       NOW(),
       $6
FROM finance_client fc
WHERE fc.client_id = $1
RETURNING (SELECT reference invoicereference FROM invoice WHERE id = invoice_id)
`

type CreatePendingInvoiceAdjustmentParams struct {
	ClientID       int32
	InvoiceID      int32
	AdjustmentType string
	Amount         int32
	Notes          string
	CreatedBy      int32
}

func (q *Queries) CreatePendingInvoiceAdjustment(ctx context.Context, arg CreatePendingInvoiceAdjustmentParams) (string, error) {
	row := q.db.QueryRow(ctx, createPendingInvoiceAdjustment,
		arg.ClientID,
		arg.InvoiceID,
		arg.AdjustmentType,
		arg.Amount,
		arg.Notes,
		arg.CreatedBy,
	)
	var invoicereference string
	err := row.Scan(&invoicereference)
	return invoicereference, err
}

const getAdjustmentForDecision = `-- name: GetAdjustmentForDecision :one
SELECT ia.amount,
       ia.adjustment_type,
       ia.finance_client_id,
       ia.invoice_id,
       i.amount - COALESCE(SUM(la.amount), 0) outstanding
FROM invoice_adjustment ia
         JOIN invoice i ON ia.invoice_id = i.id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status NOT IN ('PENDING', 'UNALLOCATED')
WHERE ia.id = $1
GROUP BY ia.amount, ia.adjustment_type, ia.finance_client_id, ia.invoice_id, i.amount
`

type GetAdjustmentForDecisionRow struct {
	Amount          int32
	AdjustmentType  string
	FinanceClientID int32
	InvoiceID       int32
	Outstanding     int32
}

func (q *Queries) GetAdjustmentForDecision(ctx context.Context, id int32) (GetAdjustmentForDecisionRow, error) {
	row := q.db.QueryRow(ctx, getAdjustmentForDecision, id)
	var i GetAdjustmentForDecisionRow
	err := row.Scan(
		&i.Amount,
		&i.AdjustmentType,
		&i.FinanceClientID,
		&i.InvoiceID,
		&i.Outstanding,
	)
	return i, err
}

const getInvoiceAdjustments = `-- name: GetInvoiceAdjustments :many
SELECT ia.id,
       i.reference AS invoice_ref,
       ia.raised_date,
       ia.adjustment_type,
       ia.amount,
       ia.notes,
       ia.status
FROM invoice_adjustment ia
         JOIN invoice i ON i.id = ia.invoice_id
         JOIN finance_client fc ON fc.id = ia.finance_client_id
WHERE fc.client_id = $1
ORDER BY ia.raised_date DESC, ia.created_at DESC
`

type GetInvoiceAdjustmentsRow struct {
	ID             int32
	InvoiceRef     string
	RaisedDate     pgtype.Date
	AdjustmentType string
	Amount         int32
	Notes          string
	Status         string
}

func (q *Queries) GetInvoiceAdjustments(ctx context.Context, clientID int32) ([]GetInvoiceAdjustmentsRow, error) {
	rows, err := q.db.Query(ctx, getInvoiceAdjustments, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetInvoiceAdjustmentsRow
	for rows.Next() {
		var i GetInvoiceAdjustmentsRow
		if err := rows.Scan(
			&i.ID,
			&i.InvoiceRef,
			&i.RaisedDate,
			&i.AdjustmentType,
			&i.Amount,
			&i.Notes,
			&i.Status,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const setAdjustmentDecision = `-- name: SetAdjustmentDecision :one
UPDATE invoice_adjustment ia
SET status     = $2,
    updated_at = NOW(),
    updated_by = $3
WHERE ia.id = $1
RETURNING ia.amount, ia.adjustment_type, ia.finance_client_id, ia.invoice_id,
    (SELECT i.amount - COALESCE(SUM(la.amount), 0) outstanding
     FROM invoice i
              LEFT JOIN ledger_allocation la
                        ON i.id = la.invoice_id AND la.status NOT IN ('PENDING', 'UNALLOCATED')
     WHERE i.id = ia.invoice_id
     GROUP BY i.amount)
`

type SetAdjustmentDecisionParams struct {
	ID        int32
	Status    string
	UpdatedBy pgtype.Int4
}

type SetAdjustmentDecisionRow struct {
	Amount          int32
	AdjustmentType  string
	FinanceClientID int32
	InvoiceID       int32
	Outstanding     int32
}

func (q *Queries) SetAdjustmentDecision(ctx context.Context, arg SetAdjustmentDecisionParams) (SetAdjustmentDecisionRow, error) {
	row := q.db.QueryRow(ctx, setAdjustmentDecision, arg.ID, arg.Status, arg.UpdatedBy)
	var i SetAdjustmentDecisionRow
	err := row.Scan(
		&i.Amount,
		&i.AdjustmentType,
		&i.FinanceClientID,
		&i.InvoiceID,
		&i.Outstanding,
	)
	return i, err
}
