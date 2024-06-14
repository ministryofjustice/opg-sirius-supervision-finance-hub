// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: invoices.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addFeeReductionToInvoices = `-- name: AddFeeReductionToInvoices :many
WITH filtered_invoices AS (
    SELECT i.id AS invoice_id, fr.id AS fee_reduction_id
    FROM invoice i
             JOIN fee_reduction fr
                  ON i.finance_client_id = fr.finance_client_id
    WHERE i.raiseddate >= (fr.datereceived - interval '6 months')
      AND i.raiseddate BETWEEN fr.startdate AND fr.enddate
      AND fr.id = $1
)
UPDATE invoice i
SET fee_reduction_id = fi.fee_reduction_id
FROM filtered_invoices fi
WHERE i.id = fi.invoice_id
returning i.id, i.person_id, i.finance_client_id, i.feetype, i.reference, i.startdate, i.enddate, i.amount, i.supervisionlevel, i.confirmeddate, i.batchnumber, i.raiseddate, i.source, i.scheduledfn14date, i.cacheddebtamount, i.createddate, i.createdby_id, i.fee_reduction_id
`

func (q *Queries) AddFeeReductionToInvoices(ctx context.Context, id int32) ([]Invoice, error) {
	rows, err := q.db.Query(ctx, addFeeReductionToInvoices, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Invoice
	for rows.Next() {
		var i Invoice
		if err := rows.Scan(
			&i.ID,
			&i.PersonID,
			&i.FinanceClientID,
			&i.Feetype,
			&i.Reference,
			&i.Startdate,
			&i.Enddate,
			&i.Amount,
			&i.Supervisionlevel,
			&i.Confirmeddate,
			&i.Batchnumber,
			&i.Raiseddate,
			&i.Source,
			&i.Scheduledfn14date,
			&i.Cacheddebtamount,
			&i.Createddate,
			&i.CreatedbyID,
			&i.FeeReductionID,
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

const getInvoiceBalance = `-- name: GetInvoiceBalance :one
SELECT i.amount initial, i.amount - COALESCE(SUM(la.amount), 0) outstanding, i.feetype
FROM invoice i
         LEFT JOIN ledger_allocation la on i.id = la.invoice_id
    AND la.status = 'ALLOCATED'
WHERE i.id = $1
group by i.amount, i.feetype
`

type GetInvoiceBalanceRow struct {
	Initial     int32
	Outstanding int32
	Feetype     string
}

func (q *Queries) GetInvoiceBalance(ctx context.Context, id int32) (GetInvoiceBalanceRow, error) {
	row := q.db.QueryRow(ctx, getInvoiceBalance, id)
	var i GetInvoiceBalanceRow
	err := row.Scan(&i.Initial, &i.Outstanding, &i.Feetype)
	return i, err
}

const getInvoices = `-- name: GetInvoices :many
SELECT i.id, i.reference, i.amount, i.raiseddate, COALESCE(SUM(la.amount), 0)::int received
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status = 'ALLOCATED'
WHERE fc.client_id = $1
GROUP BY i.id, i.raiseddate
ORDER BY i.raiseddate DESC
`

type GetInvoicesRow struct {
	ID         int32
	Reference  string
	Amount     int32
	Raiseddate pgtype.Date
	Received   int32
}

func (q *Queries) GetInvoices(ctx context.Context, clientID int32) ([]GetInvoicesRow, error) {
	rows, err := q.db.Query(ctx, getInvoices, clientID)
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
			&i.Received,
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
select la.id, la.amount, la.datetime, l.bankdate, l.type, la.status
from ledger_allocation la
         inner join ledger l on la.ledger_id = l.id
where la.invoice_id = $1
order by la.id desc
`

type GetLedgerAllocationsRow struct {
	ID       int32
	Amount   int32
	Datetime pgtype.Timestamp
	Bankdate pgtype.Date
	Type     string
	Status   string
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

const getSupervisionLevels = `-- name: GetSupervisionLevels :many
select supervisionlevel, fromdate, todate, amount
from invoice_fee_range
where invoice_id = $1
order by todate desc
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
