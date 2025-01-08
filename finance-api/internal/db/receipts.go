package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"time"
)

type Receipts struct {
	FromDate *shared.Date
	ToDate   *shared.Date
}

const ReceiptsQuery = `SELECT CONCAT(p.firstname, ' ', p.surname)                             AS "Customer Name",
       p.caserecnumber                                                                        AS "Customer number",
       fc.sop_number                                                                          AS "SOP number",
       '0470'                                                                                 AS "Entity",
       cc.code                                                                                AS "Receivables cost centre",
       cc.cost_centre_description                                                             AS "Receivables cost centre description",
       a.code                                                                                 AS "Receivables account code",
       a.account_code_description                                                             AS "Receivables account code description",
       CONCAT(tt.fee_type, i.reference)                                                       AS "Txn number",
       tt.description                                                                         AS "Txn description",
       l.bankdate                                                                             AS "Receipt date",
       l.datetime                                                                             AS "Sirius upload date",
       CASE
           WHEN la.datetime >= DATE_TRUNC('year', la.datetime) + INTERVAL '3 months'
               THEN CONCAT(EXTRACT(YEAR FROM la.datetime), '/', TO_CHAR(EXTRACT(YEAR FROM la.datetime) + 1, 'YY'))
           ELSE CONCAT(EXTRACT(YEAR FROM la.datetime) - 1, '/', TO_CHAR(EXTRACT(YEAR FROM la.datetime), 'YY'))
           END                                                                                AS "Financial Year",
       CASE WHEN la.status = 'ALLOCATED' THEN (la.amount / 100.0)::NUMERIC(10, 2) ELSE 0 END  AS "Receipt amount",
       CASE WHEN la.status <> 'UNAPPLIED' THEN (la.amount / 100.0)::NUMERIC(10, 2) ELSE 0 END AS "Amount applied",
       CASE WHEN la.status = 'UNAPPLIED' THEN (la.amount / 100.0)::NUMERIC(10, 2) ELSE 0 END  AS "Amount unapplied"
FROM supervision_finance.ledger_allocation la
         JOIN supervision_finance.ledger l ON l.id = la.ledger_id
         JOIN supervision_finance.invoice i ON i.id = la.invoice_id
         JOIN supervision_finance.finance_client fc ON fc.id = l.finance_client_id
         JOIN public.persons p ON fc.client_id = p.id
         LEFT JOIN supervision_finance.invoice_adjustment ia ON i.id = ia.invoice_id
         LEFT JOIN supervision_finance.fee_reduction fr ON fr.id = l.fee_reduction_id
         JOIN supervision_finance.transaction_type tt
              ON l.type = tt.ledger_type
         JOIN supervision_finance.account a ON tt.account_code = a.code
         JOIN supervision_finance.cost_centre cc ON cc.code = a.cost_centre
WHERE (la.status IN ('UNAPPLIED', 'REAPPLIED') OR
       (l.type IN ('MOTO CARD PAYMENT', 'ONLINE CARD PAYMENT', 'SUPERVISION BACS PAYMENT', 'OPG BACS PAYMENT') AND
        la.status = 'ALLOCATED')
    )
  AND l.datetime BETWEEN $1 AND $2;`

func (r *Receipts) GetHeaders() []string {
	return []string{
		"Customer name",
		"Customer number",
		"SOP number",
		"Entity",
		"Receivables cost centre",
		"Receivables cost centre description",
		"Receivables account code",
		"Account code description",
		"Txn number",
		"Txn type",
		"Receipt date",
		"Sirius upload date",
		"Financial Year",
		"Receipt number",
		"Receipt amount",
		"Amount applied",
		"Amount unapplied",
		"Line description (does not feature on report)",
	}
}

func (r *Receipts) GetQuery() string {
	return ReceiptsQuery
}

func (r *Receipts) GetParams() []any {
	if r.FromDate == nil {
		from := shared.NewDate("")
		r.FromDate = &from
	}

	if r.ToDate == nil {
		to := shared.Date{Time: time.Now()}
		r.ToDate = &to
	}

	return []any{r.FromDate.Time.Format("2006-01-02"), r.ToDate.Time.Format("2006-01-02")}
}
