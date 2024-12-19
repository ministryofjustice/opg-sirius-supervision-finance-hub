// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: invoices.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addInvoice = `-- name: AddInvoice :one
INSERT INTO invoice (id, person_id, finance_client_id, feetype, reference, startdate, enddate, amount, confirmeddate,
                     raiseddate, source, created_at, created_by)
VALUES (NEXTVAL('invoice_id_seq'),
        $1,
        (SELECT id FROM finance_client WHERE client_id = $1),
        $2,
        $3,
        $4,
        $5,
        $6,
        NOW(),
        $7,
        $8,
        NOW(),
        $9)
RETURNING id, person_id, finance_client_id, feetype, reference, startdate, enddate, amount, supervisionlevel, confirmeddate, batchnumber, raiseddate, source, scheduledfn14date, cacheddebtamount, created_at, created_by
`

type AddInvoiceParams struct {
	PersonID   pgtype.Int4
	Feetype    string
	Reference  string
	Startdate  pgtype.Date
	Enddate    pgtype.Date
	Amount     int32
	Raiseddate pgtype.Date
	Source     pgtype.Text
	CreatedBy  pgtype.Int4
}

func (q *Queries) AddInvoice(ctx context.Context, arg AddInvoiceParams) (Invoice, error) {
	row := q.db.QueryRow(ctx, addInvoice,
		arg.PersonID,
		arg.Feetype,
		arg.Reference,
		arg.Startdate,
		arg.Enddate,
		arg.Amount,
		arg.Raiseddate,
		arg.Source,
		arg.CreatedBy,
	)
	var i Invoice
	err := row.Scan(
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
		&i.CreatedAt,
		&i.CreatedBy,
	)
	return i, err
}

const getInvoiceBalanceDetails = `-- name: GetInvoiceBalanceDetails :one
SELECT i.amount                                                    initial,
       i.amount - COALESCE(SUM(la.amount), 0)                      outstanding,
       i.feetype,
       COALESCE((SELECT SUM(ledger_allocation.amount) FROM ledger_allocation LEFT JOIN ledger ON ledger_allocation.ledger_id = ledger.id LEFT JOIN invoice ON ledger_allocation.invoice_id = invoice.id WHERE ledger.type = 'CREDIT WRITE OFF' AND invoice.id = i.id), 0)::INT write_off_amount
FROM invoice i
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id
         LEFT JOIN ledger l ON l.id = la.ledger_id
    AND la.status NOT IN ('PENDING', 'UNALLOCATED')
WHERE i.id = $1
GROUP BY i.amount, i.feetype, i.id
`

type GetInvoiceBalanceDetailsRow struct {
	Initial        int32
	Outstanding    int32
	Feetype        string
	WriteOffAmount int32
}

func (q *Queries) GetInvoiceBalanceDetails(ctx context.Context, id int32) (GetInvoiceBalanceDetailsRow, error) {
	row := q.db.QueryRow(ctx, getInvoiceBalanceDetails, id)
	var i GetInvoiceBalanceDetailsRow
	err := row.Scan(
		&i.Initial,
		&i.Outstanding,
		&i.Feetype,
		&i.WriteOffAmount,
	)
	return i, err
}

const getInvoiceBalancesForFeeReductionRange = `-- name: GetInvoiceBalancesForFeeReductionRange :many
SELECT i.id,
       i.amount,
       COALESCE(general_fee.amount, 0)                          general_supervision_fee,
       i.amount - COALESCE(SUM(la.amount), 0) outstanding,
       i.feetype
FROM invoice i
         JOIN fee_reduction fr ON i.finance_client_id = fr.finance_client_id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id
         LEFT JOIN ledger l ON l.id = la.ledger_id
         LEFT JOIN LATERAL (
             SELECT SUM(ifr.amount) AS amount
             FROM invoice_fee_range ifr
             WHERE ifr.invoice_id = i.id
             AND ifr.supervisionlevel = 'GENERAL'
         ) general_fee ON TRUE
WHERE i.raiseddate >= (fr.datereceived - INTERVAL '6 months')
  AND i.raiseddate BETWEEN fr.startdate AND fr.enddate
  AND fr.id = $1
GROUP BY i.id, general_fee.amount
`

type GetInvoiceBalancesForFeeReductionRangeRow struct {
	ID                    int32
	Amount                int32
	GeneralSupervisionFee int64
	Outstanding           int32
	Feetype               string
}

func (q *Queries) GetInvoiceBalancesForFeeReductionRange(ctx context.Context, id int32) ([]GetInvoiceBalancesForFeeReductionRangeRow, error) {
	rows, err := q.db.Query(ctx, getInvoiceBalancesForFeeReductionRange, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetInvoiceBalancesForFeeReductionRangeRow
	for rows.Next() {
		var i GetInvoiceBalancesForFeeReductionRangeRow
		if err := rows.Scan(
			&i.ID,
			&i.Amount,
			&i.GeneralSupervisionFee,
			&i.Outstanding,
			&i.Feetype,
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

const getInvoiceCounter = `-- name: GetInvoiceCounter :one
INSERT INTO counter (id, key, counter)
VALUES (NEXTVAL('counter_id_seq'), $1, 1)
ON CONFLICT (key) DO UPDATE
    SET counter = counter.counter + 1
RETURNING counter::VARCHAR
`

func (q *Queries) GetInvoiceCounter(ctx context.Context, key string) (string, error) {
	row := q.db.QueryRow(ctx, getInvoiceCounter, key)
	var counter string
	err := row.Scan(&counter)
	return counter, err
}

const getInvoices = `-- name: GetInvoices :many
SELECT i.id,
       i.raiseddate,
       i.reference,
       i.amount,
       COALESCE(SUM(la.amount), 0)::INT    received,
       COALESCE(MAX(fr.type), '')::VARCHAR fee_reduction_type
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status NOT IN ('PENDING', 'UNALLOCATED')
         LEFT JOIN ledger l ON la.ledger_id = l.id
         LEFT JOIN fee_reduction fr ON l.fee_reduction_id = fr.id
WHERE fc.client_id = $1
GROUP BY i.id, i.raiseddate
ORDER BY i.raiseddate DESC
`

type GetInvoicesRow struct {
	ID               int32
	Raiseddate       pgtype.Date
	Reference        string
	Amount           int32
	Received         int32
	FeeReductionType string
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
			&i.Raiseddate,
			&i.Reference,
			&i.Amount,
			&i.Received,
			&i.FeeReductionType,
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

const getInvoicesForCourtRef = `-- name: GetInvoicesForCourtRef :many
SELECT i.id,
       (i.amount - COALESCE(SUM(la.amount), 0)::INT) outstanding
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status NOT IN ('PENDING', 'UNALLOCATED')
         LEFT JOIN ledger l ON la.ledger_id = l.id
         LEFT JOIN fee_reduction fr ON l.fee_reduction_id = fr.id
WHERE fc.court_ref = $1
GROUP BY i.id, i.raiseddate
HAVING (i.amount - COALESCE(SUM(la.amount), 0)::INT) > 0
ORDER BY i.raiseddate ASC
`

type GetInvoicesForCourtRefRow struct {
	ID          int32
	Outstanding int32
}

func (q *Queries) GetInvoicesForCourtRef(ctx context.Context, courtRef pgtype.Text) ([]GetInvoicesForCourtRefRow, error) {
	rows, err := q.db.Query(ctx, getInvoicesForCourtRef, courtRef)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetInvoicesForCourtRefRow
	for rows.Next() {
		var i GetInvoicesForCourtRefRow
		if err := rows.Scan(&i.ID, &i.Outstanding); err != nil {
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
WITH allocations AS (SELECT la.invoice_id,
                            la.amount,
                            COALESCE(l.bankdate, la.datetime) AS raised_date,
                            l.type,
                            la.status,
                            la.datetime AS created_at,
                            la.id AS ledger_allocation_id
                     FROM ledger_allocation la
                              JOIN ledger l ON la.ledger_id = l.id
                     WHERE la.invoice_id = ANY ($1::INT[])
                     UNION
                     SELECT ia.invoice_id, ia.amount, ia.raised_date, ia.adjustment_type, ia.status, ia.created_at, ia.id
                     FROM invoice_adjustment ia
                     WHERE ia.status = 'PENDING'
                       AND ia.invoice_id = ANY ($1::INT[]))
SELECT invoice_id, amount, raised_date, type, status, created_at, ledger_allocation_id
FROM allocations
ORDER BY raised_date DESC, created_at DESC, status DESC, ledger_allocation_id ASC
`

type GetLedgerAllocationsRow struct {
	InvoiceID          pgtype.Int4
	Amount             int32
	RaisedDate         pgtype.Date
	Type               string
	Status             string
	CreatedAt          pgtype.Timestamp
	LedgerAllocationID int32
}

func (q *Queries) GetLedgerAllocations(ctx context.Context, dollar_1 []int32) ([]GetLedgerAllocationsRow, error) {
	rows, err := q.db.Query(ctx, getLedgerAllocations, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetLedgerAllocationsRow
	for rows.Next() {
		var i GetLedgerAllocationsRow
		if err := rows.Scan(
			&i.InvoiceID,
			&i.Amount,
			&i.RaisedDate,
			&i.Type,
			&i.Status,
			&i.CreatedAt,
			&i.LedgerAllocationID,
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
SELECT invoice_id, supervisionlevel, fromdate, todate, amount
FROM invoice_fee_range
WHERE invoice_id = ANY ($1::INT[])
ORDER BY todate DESC
`

type GetSupervisionLevelsRow struct {
	InvoiceID        pgtype.Int4
	Supervisionlevel string
	Fromdate         pgtype.Date
	Todate           pgtype.Date
	Amount           int32
}

func (q *Queries) GetSupervisionLevels(ctx context.Context, dollar_1 []int32) ([]GetSupervisionLevelsRow, error) {
	rows, err := q.db.Query(ctx, getSupervisionLevels, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetSupervisionLevelsRow
	for rows.Next() {
		var i GetSupervisionLevelsRow
		if err := rows.Scan(
			&i.InvoiceID,
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
