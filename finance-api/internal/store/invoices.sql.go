// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: invoices.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getInvoices = `-- name: GetInvoices :many
SELECT id, reference, amount, raiseddate, cacheddebtamount FROM invoice WHERE finance_client_id = $1 order by raiseddate desc
`

type GetInvoicesRow struct {
	ID               int32
	Reference        string
	Amount           int32
	Raiseddate       pgtype.Date
	Cacheddebtamount pgtype.Int4
}

func (q *Queries) GetInvoices(ctx context.Context, financeClientID pgtype.Int4) ([]GetInvoicesRow, error) {
	rows, err := q.db.Query(ctx, getInvoices, financeClientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetInvoicesRow
	for rows.Next() {
		var i GetInvoicesRow
		if err := rows.Scan(
			&i.ID,
			&i.Reference,
			&i.Amount,
			&i.Raiseddate,
			&i.Cacheddebtamount,
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

const getLedgerAllocations = `-- name: GetLedgerAllocations :many
select la.id, la.amount, la.datetime, l.bankdate, l.type from ledger_allocation la inner join ledger l on la.ledger_id = l.id where la.invoice_id = $1 order by la.id desc
`

type GetLedgerAllocationsRow struct {
	ID       int32
	Amount   int32
	Datetime pgtype.Timestamp
	Bankdate pgtype.Date
	Type     string
}

func (q *Queries) GetLedgerAllocations(ctx context.Context, invoiceID pgtype.Int4) ([]GetLedgerAllocationsRow, error) {
	rows, err := q.db.Query(ctx, getLedgerAllocations, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetLedgerAllocationsRow
	for rows.Next() {
		var i GetLedgerAllocationsRow
		if err := rows.Scan(
			&i.ID,
			&i.Amount,
			&i.Datetime,
			&i.Bankdate,
			&i.Type,
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

const getSupervisionLevels = `-- name: GetSupervisionLevels :many
select supervisionlevel, fromdate, todate, amount from invoice_fee_range where invoice_id = $1 order by todate desc
`

type GetSupervisionLevelsRow struct {
	Supervisionlevel string
	Fromdate         pgtype.Date
	Todate           pgtype.Date
	Amount           int32
}

func (q *Queries) GetSupervisionLevels(ctx context.Context, invoiceID pgtype.Int4) ([]GetSupervisionLevelsRow, error) {
	rows, err := q.db.Query(ctx, getSupervisionLevels, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetSupervisionLevelsRow
	for rows.Next() {
		var i GetSupervisionLevelsRow
		if err := rows.Scan(
			&i.Supervisionlevel,
			&i.Fromdate,
			&i.Todate,
			&i.Amount,
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