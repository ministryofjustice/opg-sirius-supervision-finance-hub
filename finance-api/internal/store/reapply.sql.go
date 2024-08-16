// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: reapply.sql

package store

import (
	"context"
)

const getCreditBalanceAndOldestOpenInvoice = `-- name: GetCreditBalanceAndOldestOpenInvoice :one
SELECT (SELECT ABS(COALESCE(SUM(la.amount), 0))
        FROM finance_client fc
                 LEFT JOIN ledger l ON fc.id = l.finance_client_id
                 LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
        WHERE fc.client_id = $1
          AND la.status IN ('UNAPPLIED', 'REAPPLIED'))::int AS credit,
       i.id AS invoice_id,
       i.amount AS invoiceAmount,
       i.amount - COALESCE(SUM(la.amount), 0) AS outstanding
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status NOT IN ('PENDING', 'UNALLOCATED')
         LEFT JOIN ledger l ON la.ledger_id = l.id
WHERE fc.client_id = $1
GROUP BY i.id, i.raiseddate, i.amount
HAVING COALESCE(SUM(la.amount), 0) < i.amount
ORDER BY i.raiseddate LIMIT 1
`

type GetCreditBalanceAndOldestOpenInvoiceRow struct {
	Credit        int32
	InvoiceID     int32
	Invoiceamount int32
	Outstanding   int32
}

func (q *Queries) GetCreditBalanceAndOldestOpenInvoice(ctx context.Context, clientID int32) (GetCreditBalanceAndOldestOpenInvoiceRow, error) {
	row := q.db.QueryRow(ctx, getCreditBalanceAndOldestOpenInvoice, clientID)
	var i GetCreditBalanceAndOldestOpenInvoiceRow
	err := row.Scan(
		&i.Credit,
		&i.InvoiceID,
		&i.Invoiceamount,
		&i.Outstanding,
	)
	return i, err
}
