// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: fee_reductions.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addFeeReduction = `-- name: AddFeeReduction :one
INSERT INTO fee_reduction (id,
                           finance_client_id,
                           type,
                           startdate,
                           enddate,
                           notes,
                           datereceived,
                           created_by,
                           created_at)
VALUES (NEXTVAL('fee_reduction_id_seq'::REGCLASS),
        (SELECT id FROM finance_client WHERE client_id = $1), $2, $3::DATE, $4::DATE, $5, $6, $7, now())
RETURNING id, finance_client_id, type, evidencetype, startdate, enddate, notes, deleted, datereceived, created_at, created_by, cancelled_at, cancelled_by, cancellation_reason
`

type AddFeeReductionParams struct {
	ClientID     int32
	Type         string
	StartDate    pgtype.Date
	EndDate      pgtype.Date
	Notes        string
	DateReceived pgtype.Date
	CreatedBy    pgtype.Int4
}

func (q *Queries) AddFeeReduction(ctx context.Context, arg AddFeeReductionParams) (FeeReduction, error) {
	row := q.db.QueryRow(ctx, addFeeReduction,
		arg.ClientID,
		arg.Type,
		arg.StartDate,
		arg.EndDate,
		arg.Notes,
		arg.DateReceived,
		arg.CreatedBy,
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
		&i.CreatedAt,
		&i.CreatedBy,
		&i.CancelledAt,
		&i.CancelledBy,
		&i.CancellationReason,
	)
	return i, err
}

const cancelFeeReduction = `-- name: CancelFeeReduction :one
UPDATE fee_reduction
SET deleted = TRUE,  cancelled_by = $2, cancelled_at = now(), cancellation_reason = $3
WHERE id = $1
RETURNING id, finance_client_id, type, evidencetype, startdate, enddate, notes, deleted, datereceived, created_at, created_by, cancelled_at, cancelled_by, cancellation_reason
`

type CancelFeeReductionParams struct {
	ID                 int32
	CancelledBy        pgtype.Int4
	CancellationReason pgtype.Text
}

func (q *Queries) CancelFeeReduction(ctx context.Context, arg CancelFeeReductionParams) (FeeReduction, error) {
	row := q.db.QueryRow(ctx, cancelFeeReduction, arg.ID, arg.CancelledBy, arg.CancellationReason)
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
		&i.CreatedAt,
		&i.CreatedBy,
		&i.CancelledAt,
		&i.CancelledBy,
		&i.CancellationReason,
	)
	return i, err
}

const countOverlappingFeeReduction = `-- name: CountOverlappingFeeReduction :one
SELECT COUNT(*)
FROM fee_reduction fr
         INNER JOIN finance_client fc ON fc.id = fr.finance_client_id
WHERE fc.client_id = $1
  AND fr.deleted = FALSE
  AND (fr.startdate::DATE, fr.enddate::DATE) OVERLAPS ($2::DATE, $3::DATE)
`

type CountOverlappingFeeReductionParams struct {
	ClientID  int32
	StartDate pgtype.Date
	EndDate   pgtype.Date
}

func (q *Queries) CountOverlappingFeeReduction(ctx context.Context, arg CountOverlappingFeeReductionParams) (int64, error) {
	row := q.db.QueryRow(ctx, countOverlappingFeeReduction, arg.ClientID, arg.StartDate, arg.EndDate)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getFeeReductionForDate = `-- name: GetFeeReductionForDate :one
SELECT fr.id, fr.type
FROM fee_reduction fr
         JOIN finance_client fc ON fr.finance_client_id = fc.id
WHERE $1::DATE >= (fr.datereceived - INTERVAL '6 months')
  AND $1::DATE BETWEEN fr.startdate::DATE AND fr.enddate::DATE
  AND fr.deleted = FALSE
  AND fc.client_id = $2
`

type GetFeeReductionForDateParams struct {
	DateReceived pgtype.Date
	ClientID     int32
}

type GetFeeReductionForDateRow struct {
	ID   int32
	Type string
}

func (q *Queries) GetFeeReductionForDate(ctx context.Context, arg GetFeeReductionForDateParams) (GetFeeReductionForDateRow, error) {
	row := q.db.QueryRow(ctx, getFeeReductionForDate, arg.DateReceived, arg.ClientID)
	var i GetFeeReductionForDateRow
	err := row.Scan(&i.ID, &i.Type)
	return i, err
}

const getFeeReductions = `-- name: GetFeeReductions :many
SELECT fr.id,
       finance_client_id,
       type,
       startdate,
       enddate,
       datereceived,
       notes,
       deleted
FROM fee_reduction fr
         INNER JOIN finance_client fc ON fc.id = fr.finance_client_id
WHERE fc.client_id = $1
ORDER BY enddate DESC, deleted
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
