package db

import (
	"github.com/jackc/pgx/v5"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"time"
)

type InvoiceAdjustments struct {
	FromDate   *shared.Date
	ToDate     *shared.Date
	GoLiveDate time.Time
}

const InvoiceAdjustmentsQuery = `
SELECT CONCAT(p.firstname, ' ', p.surname)               AS "Customer Name",
       p.caserecnumber                                   AS "Customer number",
       fc.sop_number                                     AS "SOP number",
       '="0470"'                                         AS "Entity",
       cc.code                                           AS "Revenue cost centre",
       cc.cost_centre_description                        AS "Revenue cost centre description",
       a.code                                            AS "Revenue account code",
       a.account_code_description                        AS "Revenue account code description",
       CONCAT(tt.fee_type, i.reference)                  AS "Txn number",
       tt.description                                    AS "Txn description",
       CASE
           WHEN fr.startdate IS NOT NULL
               THEN CONCAT(EXTRACT(YEAR FROM AGE(fr.enddate, fr.startdate)), ' year')
           ELSE ''
           END                                           AS "Remission/Exemption award term",
       CASE
           WHEN la.datetime >= DATE_TRUNC('year', la.datetime) + INTERVAL '3 months'
               THEN CONCAT(TO_CHAR(la.datetime, 'YY'), '/', TO_CHAR((la.datetime + INTERVAL '1 YEAR'), 'YY'))
           ELSE CONCAT(TO_CHAR((la.datetime - INTERVAL '1 YEAR'), 'YY'), '/', TO_CHAR(la.datetime, 'YY'))
           END                                           AS "Financial Year",
       TO_CHAR(la.datetime, 'YYYY-MM-DD')                AS "Approved date",
       (la.amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS "Adjustment amount",
       COALESCE(ia.notes, fr.notes)                      AS "Reason for adjustment"
FROM supervision_finance.ledger_allocation la
         JOIN supervision_finance.ledger l ON l.id = la.ledger_id
         JOIN supervision_finance.invoice i ON i.id = la.invoice_id
         JOIN supervision_finance.finance_client fc ON fc.id = l.finance_client_id
         JOIN public.persons p ON fc.client_id = p.id
         LEFT JOIN supervision_finance.invoice_adjustment ia ON i.id = ia.invoice_id
         LEFT JOIN supervision_finance.fee_reduction fr ON fr.id = l.fee_reduction_id
         INNER JOIN LATERAL (
    SELECT CASE
               WHEN i.feetype IN ('AD', 'GA', 'GS', 'GT') THEN i.feetype
               ELSE (SELECT COALESCE(ifr.supervisionlevel, '')
                     FROM supervision_finance.invoice_fee_range ifr
                     WHERE ifr.invoice_id = i.id
                     ORDER BY id DESC
                     LIMIT 1) END AS supervision_level
    ) sl ON TRUE
         JOIN supervision_finance.transaction_type tt
              ON l.type = tt.ledger_type AND sl.supervision_level = tt.supervision_level
         JOIN supervision_finance.account a ON tt.account_code = a.code
         JOIN supervision_finance.cost_centre cc ON cc.code = a.cost_centre
WHERE la.status = 'ALLOCATED'
  AND ((ia.status = 'APPROVED' AND ia.adjustment_type NOT IN ('CREDIT WRITE OFF', 'WRITE OFF REVERSAL')) OR
       fr.id IS NOT NULL)
  AND l.datetime::DATE BETWEEN $1 AND $2;
`

func (i *InvoiceAdjustments) GetHeaders() []string {
	return []string{
		"Customer Name",
		"Customer number",
		"SOP number",
		"Entity",
		"Revenue cost centre",
		"Revenue cost centre description",
		"Revenue account code",
		"Revenue account descriptions",
		"Txn number and type",
		"Txn description",
		"Remission/exemption term",
		"Financial Year",
		"Approved date",
		"Adjustment amount",
		"Reason for adjustment",
	}
}

func (i *InvoiceAdjustments) GetQuery() string {
	return InvoiceAdjustmentsQuery
}

func (i *InvoiceAdjustments) GetParams() []any {
	var (
		from, to time.Time
	)

	if i.FromDate == nil {
		from = i.GoLiveDate
	} else {
		from = i.FromDate.Time
	}

	if i.ToDate == nil {
		to = time.Now()
	} else {
		to = i.ToDate.Time
	}

	return []any{from.Format("2006-01-02"), to.Format("2006-01-02")}
}

func (i *InvoiceAdjustments) GetCallback() func(row pgx.CollectableRow) ([]string, error) {
	return RowToStringMap
}
