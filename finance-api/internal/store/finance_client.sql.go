// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: finance_client.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getAccountInformation = `-- name: GetAccountInformation :one
SELECT cacheddebtamount, cachedcreditamount, payment_method FROM finance_client WHERE client_id = $1
`

type GetAccountInformationRow struct {
	Cacheddebtamount   pgtype.Int4
	Cachedcreditamount pgtype.Int4
	PaymentMethod      string
}

func (q *Queries) GetAccountInformation(ctx context.Context, clientID int32) (GetAccountInformationRow, error) {
	row := q.db.QueryRow(ctx, getAccountInformation, clientID)
	var i GetAccountInformationRow
	err := row.Scan(&i.Cacheddebtamount, &i.Cachedcreditamount, &i.PaymentMethod)
	return i, err
}
