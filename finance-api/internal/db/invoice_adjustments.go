package db

import (
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"os"
	"time"
)

type InvoiceAdjustments struct {
	FromDate *shared.Date
	ToDate   *shared.Date
}

const InvoiceAdjustmentsQuery = `SELECT CONCAT(p.firstname, ' ', p.surname)                               AS "Customer Name",
   p.caserecnumber                                                   AS "Customer number",
   fc.sop_number                                                     AS "SOP number",
   '="0470"'                                                            AS "Entity",
   cc.code                                                           AS "Revenue cost centre",
   cc.cost_centre_description                                        AS "Revenue cost centre description",
   a.code                                                            AS "Revenue account code",
   a.account_code_description                                        AS "Revenue account code description",
   CONCAT(tt.fee_type, i.reference)                                  AS "Txn number",
   tt.description                                                    AS "Txn description",
   CASE WHEN fr.startdate IS NOT NULL AND fr.enddate IS NOT NULL THEN CONCAT(EXTRACT(YEAR FROM AGE(fr.enddate, fr.startdate)), ' year') END AS "Remission/Exemption award term",
   CASE
       WHEN la.datetime >= DATE_TRUNC('year', la.datetime) + INTERVAL '3 months'
           THEN CONCAT(TO_CHAR(la.datetime, 'YY'), '/', TO_CHAR((la.datetime + INTERVAL '1 YEAR'), 'YY'))
       ELSE CONCAT(TO_CHAR((la.datetime - INTERVAL '1 YEAR'), 'YY'), '/', TO_CHAR(la.datetime, 'YY'))
       END                                                           AS "Financial Year",
   la.datetime                                                       AS "Approved date",
   (la.amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255)                               AS "Adjustment amount",
   COALESCE(ia.notes, fr.notes)                                      AS "Reason for adjustment"
FROM supervision_finance.ledger_allocation la
     JOIN supervision_finance.ledger l on l.id = la.ledger_id
     JOIN supervision_finance.invoice i on i.id = la.invoice_id
     JOIN supervision_finance.finance_client fc on fc.id = l.finance_client_id
     JOIN public.persons p ON fc.client_id = p.id
     LEFT JOIN supervision_finance.invoice_adjustment ia on i.id = ia.invoice_id
     LEFT JOIN supervision_finance.fee_reduction fr ON fr.id = l.fee_reduction_id
     LEFT JOIN LATERAL (
SELECT CASE WHEN i.feetype = 'AD' THEN 'AD' ELSE COALESCE(ifr.supervisionlevel, '') END AS supervision_level
FROM supervision_finance.invoice_fee_range ifr
WHERE ifr.invoice_id = i.id
ORDER BY id
LIMIT 1
) sl ON TRUE
     JOIN supervision_finance.transaction_type tt
          ON l.type = tt.ledger_type AND sl.supervision_level = tt.supervision_level
     JOIN supervision_finance.account a ON tt.account_code = a.code
     JOIN supervision_finance.cost_centre cc ON cc.code = a.cost_centre
WHERE la.status = 'ALLOCATED'
AND ((ia.status = 'APPROVED' AND ia.adjustment_type NOT IN ('CREDIT WRITE OFF', 'WRITE OFF REVERSAL')) OR
   fr.id IS NOT NULL)
AND l.datetime BETWEEN $1 AND $2;`

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
	if i.FromDate == nil {
		from := shared.NewDate(os.Getenv("FINANCE_HUB_LIVE_DATE"))
		i.FromDate = &from
	}

	if i.ToDate == nil {
		to := shared.Date{Time: time.Now()}
		i.ToDate = &to
	}

	from := fmt.Sprintf("%s 00:00:00", i.FromDate.Time.Format("2006-01-02"))
	to := fmt.Sprintf("%s 23:59:59", i.ToDate.Time.Format("2006-01-02"))

	return []any{from, to}
}
