// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: pending_invoice_adjustments.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getInvoiceAdjustments = `-- name: GetInvoiceAdjustments :many
SELECT l.id,
       i.reference AS invoice_ref,
       l.datetime  AS raised_date,
       l.type,
       l.amount,
       l.notes,
       l.status
FROM supervision_finance.ledger l
         INNER JOIN supervision_finance.ledger_allocation lc ON lc.ledger_id = l.id
         INNER JOIN supervision_finance.invoice i ON i.id = lc.invoice_id
         INNER JOIN supervision_finance.finance_client fc ON fc.id = i.finance_client_id
WHERE fc.client_id = $1
  AND l.type IN ('CREDIT MEMO', 'CREDIT WRITE OFF', 'DEBIT MEMO')
ORDER BY l.datetime DESC
`

type GetInvoiceAdjustmentsRow struct {
	ID         int32
	InvoiceRef string
	RaisedDate pgtype.Timestamp
	Type       string
	Amount     int32
	Notes      pgtype.Text
	Status     string
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
			&i.Type,
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
