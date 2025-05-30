// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: adhoc_script.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getNegativeInvoices = `-- name: GetNegativeInvoices :many
WITH total_debt AS (select i.id, fc.court_ref, i.reference, i.amount as invoiceamount, SUM(la.amount) as laamount
                    from invoice  i
                             inner join ledger_allocation la on la.invoice_id = i.id
                             inner join ledger l on l.id = la.ledger_id
                             inner join finance_client fc on i.finance_client_id = fc.id
                    where la.status NOT IN ('PENDING', 'UNALLOCATED') and l.status = 'CONFIRMED'
                    group by i.id, i.reference, fc.court_ref),
     negInvoices AS (select td.id, td.reference, td.court_ref, sum(invoiceamount - laamount) as ledgerallocationamountneeded, i.person_id
                     from total_debt td
                              inner join invoice i on i.id = td.id
                     group by td.id, td.reference, td.court_ref, i.person_id
                     HAVING sum(invoiceamount - laamount) < 0)
select distinct on (ni.id) l.id as ledgerId, l.type, ni.reference, ni.id as invoiceId, ni.court_ref, ni.ledgerallocationamountneeded, ni.person_id
from negInvoices ni
         inner join ledger_allocation la on ni.id = la.invoice_id
         inner join ledger l on la.ledger_id = l.id
order by ni.id, l.id desc
`

type GetNegativeInvoicesRow struct {
	Ledgerid                     int32
	Type                         string
	Reference                    string
	Invoiceid                    int32
	CourtRef                     pgtype.Text
	Ledgerallocationamountneeded int64
	PersonID                     pgtype.Int4
}

func (q *Queries) GetNegativeInvoices(ctx context.Context) ([]GetNegativeInvoicesRow, error) {
	rows, err := q.db.Query(ctx, getNegativeInvoices)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetNegativeInvoicesRow
	for rows.Next() {
		var i GetNegativeInvoicesRow
		if err := rows.Scan(
			&i.Ledgerid,
			&i.Type,
			&i.Reference,
			&i.Invoiceid,
			&i.CourtRef,
			&i.Ledgerallocationamountneeded,
			&i.PersonID,
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
