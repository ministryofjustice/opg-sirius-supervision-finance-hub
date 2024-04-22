// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: fee_reductions.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addFeeReduction = `-- name: AddFeeReduction :one
insert into fee_reduction (finance_client_id,
                           type,
                           evidencetype,
                           startdate,
                           enddate,
                           notes,
                           deleted,
                           datereceived) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id, finance_client_id, type, evidencetype, startdate, enddate, notes, deleted, datereceived
`

type AddFeeReductionParams struct {
	FinanceClientID pgtype.Int4
	Type            string
	Evidencetype    pgtype.Text
	Startdate       pgtype.Date
	Enddate         pgtype.Date
	Notes           string
	Deleted         bool
	Datereceived    pgtype.Date
}

func (q *Queries) AddFeeReduction(ctx context.Context, arg AddFeeReductionParams) (FeeReduction, error) {
	row := q.db.QueryRow(ctx, addFeeReduction,
		arg.FinanceClientID,
		arg.Type,
		arg.Evidencetype,
		arg.Startdate,
		arg.Enddate,
		arg.Notes,
		arg.Deleted,
		arg.Datereceived,
	)
	var i FeeReduction
	err := row.Scan(
		&i.ID,
		&i.FinanceClientID,
		&i.Type,
		&i.Evidencetype,
		&i.Startdate,
		&i.Enddate,
		&i.Notes,
		&i.Deleted,
		&i.Datereceived,
	)
	return i, err
}

const getFeeReductions = `-- name: GetFeeReductions :many
select fr.id,
       finance_client_id,
       type,
       startdate,
       enddate,
       datereceived,
       notes,
       deleted
from fee_reduction fr
         inner join finance_client fc on fc.id = fr.finance_client_id
where fc.client_id = $1
order by enddate desc, deleted
`

type GetFeeReductionsRow struct {
	ID              int32
	FinanceClientID pgtype.Int4
	Type            string
	Startdate       pgtype.Date
	Enddate         pgtype.Date
	Datereceived    pgtype.Date
	Notes           string
	Deleted         bool
}

func (q *Queries) GetFeeReductions(ctx context.Context, clientID int32) ([]GetFeeReductionsRow, error) {
	rows, err := q.db.Query(ctx, getFeeReductions, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeeReductionsRow
	for rows.Next() {
		var i GetFeeReductionsRow
		if err := rows.Scan(
			&i.ID,
			&i.FinanceClientID,
			&i.Type,
			&i.Startdate,
			&i.Enddate,
			&i.Datereceived,
			&i.Notes,
			&i.Deleted,
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
