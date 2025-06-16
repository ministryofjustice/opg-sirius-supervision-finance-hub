package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"time"
)

type Receipts struct {
	ReportQuery
	ReceiptsParams
}

type ReceiptsParams struct {
	FromDate *shared.Date
	ToDate   *shared.Date
}

func NewReceipts(params ReceiptsParams) ReportQuery {
	return &Receipts{
		ReportQuery:    NewReportQuery(ReceiptsQuery),
		ReceiptsParams: params,
	}
}

const ReceiptsQuery = `SELECT CONCAT(p.firstname, ' ', p.surname)                             AS "Customer Name",
       p.caserecnumber                                                                        AS "Customer number",
       fc.sop_number                                                                          AS "SOP number",
       '0470'                                                                                 AS "Entity",
       cc.code                                                                                AS "Receivables cost centre",
       cc.cost_centre_description                                                             AS "Receivables cost centre description",
       a.code                                                                                 AS "Receivables account code",
       a.account_code_description                                                             AS "Receivables account code description",
       CONCAT(ft.fee_type, COALESCE(i.reference, p.caserecnumber))                            AS "Txn number",
       ft.description                                                                         AS "Txn description",
       CASE WHEN l.bankdate IS NOT NULL THEN TO_CHAR(l.bankdate, 'YYYY-MM-DD') ELSE '' END    AS "Receipt date",
       TO_CHAR(l.datetime, 'YYYY-MM-DD')                                                      AS "Sirius upload date",
       CASE
           WHEN la.datetime >= DATE_TRUNC('year', la.datetime) + INTERVAL '3 months'
               THEN CONCAT(EXTRACT(YEAR FROM la.datetime), '/', TO_CHAR(la.datetime + INTERVAL '1 year', 'YY'))
           ELSE CONCAT(EXTRACT(YEAR FROM la.datetime - INTERVAL '1 year'), '/', TO_CHAR(la.datetime, 'YY'))
           END                                                                                AS "Financial Year",
       CASE WHEN la.status = 'ALLOCATED' OR (la.status = 'UNAPPLIED' AND la.invoice_id IS NULL) 
           THEN ((l.amount / 100.0)::NUMERIC(10, 2))::VARCHAR(255) ELSE '0.00' END  		  AS "Receipt amount",
       CASE WHEN la.status <> 'UNAPPLIED' THEN ((la.amount / 100.0)::NUMERIC(10, 2))::VARCHAR(255) ELSE '0.00' END AS "Amount applied",
       CASE WHEN la.status = 'UNAPPLIED' THEN ((-la.amount / 100.0)::NUMERIC(10, 2))::VARCHAR(255) ELSE '0.00' END  AS "Amount unapplied"       
FROM supervision_finance.ledger_allocation la
         JOIN supervision_finance.ledger l ON l.id = la.ledger_id
         LEFT JOIN supervision_finance.invoice i ON i.id = la.invoice_id
         JOIN supervision_finance.finance_client fc ON fc.id = l.finance_client_id
         JOIN public.persons p ON fc.client_id = p.id
    	 JOIN supervision_finance.transaction_type ft ON CASE WHEN la.status = 'UNAPPLIED' AND la.invoice_id IS NOT NULL THEN ft.fee_type = 'UA' ELSE l.type = ft.ledger_type END
         JOIN supervision_finance.transaction_type tt ON 
             CASE WHEN la.status = 'UNAPPLIED' THEN 
                  CASE WHEN la.invoice_id IS NULL THEN  tt.fee_type = 'OP' ELSE tt.fee_type = 'UA' END 
			 ELSE l.type = tt.ledger_type END
         JOIN supervision_finance.account a ON tt.account_code = a.code
         JOIN supervision_finance.cost_centre cc ON cc.code = a.cost_centre
WHERE (la.status IN ('UNAPPLIED', 'REAPPLIED') OR
       (l.type IN ('MOTO CARD PAYMENT', 'ONLINE CARD PAYMENT', 'SUPERVISION BACS PAYMENT', 'OPG BACS PAYMENT') AND
        la.status = 'ALLOCATED'))
  AND l.datetime::DATE BETWEEN $1 AND $2
ORDER BY la.id;`

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
		"Receipt amount",
		"Amount applied",
		"Amount unapplied",
	}
}

func (r *Receipts) GetParams() []any {
	var (
		from, to time.Time
	)

	if r.FromDate == nil {
		from = time.Time{}
	} else {
		from = r.FromDate.Time
	}

	if r.ToDate == nil {
		to = time.Now()
	} else {
		to = r.ToDate.Time
	}

	return []any{from.Format("2006-01-02"), to.Format("2006-01-02")}
}
