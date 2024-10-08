// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: invoice_fee_range.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addInvoiceRange = `-- name: AddInvoiceRange :exec
INSERT INTO invoice_fee_range (id, invoice_id, supervisionlevel, fromdate, todate, amount)
VALUES (nextval('invoice_fee_range_id_seq'),
        $1,
        $2,
        $3,
        $4,
        $5)
`

type AddInvoiceRangeParams struct {
	InvoiceID        pgtype.Int4
	Supervisionlevel string
	Fromdate         pgtype.Date
	Todate           pgtype.Date
	Amount           int32
}

func (q *Queries) AddInvoiceRange(ctx context.Context, arg AddInvoiceRangeParams) error {
	_, err := q.db.Exec(ctx, addInvoiceRange,
		arg.InvoiceID,
		arg.Supervisionlevel,
		arg.Fromdate,
		arg.Todate,
		arg.Amount,
	)
	return err
}
