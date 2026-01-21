package db

import (
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

// AgedDebt generates a report of all outstanding debt as of a given date. If no date is provided, the current date is used.
// Outstanding debt is defined as invoices that were raised by the provided date but not fully paid by that date.
// This means invoices raised after the provided date are not included, even if they are unpaid, and invoices that were
// fully paid after the provided date will appear as outstanding.
// The report also calculates the age of the debt based on the due date (30 days after the raised date) and categorises
// the debt into ageing buckets (current, 0-1 years, 1-2 years, etc.).
// This report is a snapshot of debt as of the specified date and as a result, the ageing buckets reflect the age of the
// debt at that time, not the current age of the debt, or if it is indeed still debt.
type AgedDebt struct {
	ReportQuery
	AgedDebtInput
}

type AgedDebtInput struct {
	ToDate *shared.Date
	Today  time.Time
}

func NewAgedDebt(input AgedDebtInput) ReportQuery {
	return &AgedDebt{
		ReportQuery:   NewReportQuery(AgedDebtQuery),
		AgedDebtInput: input,
	}
}

const AgedDebtQuery = `WITH receipt_transactions_types AS (
    SELECT DISTINCT(ledger_type) FROM supervision_finance.transaction_type WHERE ledger_type NOT LIKE '' AND is_receipt IS TRUE
),
outstanding_invoices AS (SELECT i.id,
                                     i.finance_client_id,
                                     i.feetype,
                                     CASE 
                                         WHEN i.feetype IN ('AD', 'GA', 'GS', 'GT') THEN i.feetype
                                    	 ELSE COALESCE(sl.supervision_level, '')    
									 END AS supervision_level,
                                     i.reference,
                                     i.raiseddate,
                                     i.raiseddate + '30 days'::INTERVAL AS due_date,
                                     ((i.amount / 100.0)::NUMERIC(10, 2))::VARCHAR(255) AS amount,
                                     (((i.amount - COALESCE(transactions.received, 0)) / 100.00)::NUMERIC(10, 2))::VARCHAR(255) AS outstanding,
									 DATE_PART('year', AGE($1::DATE, (i.raiseddate + '30 days'::INTERVAL))) + 
									 DATE_PART('month', AGE($1::DATE, (i.raiseddate + '30 days'::INTERVAL))) / 12.0 AS age
                              FROM supervision_finance.invoice i
									   LEFT JOIN LATERAL (
									  SELECT SUM(la.amount) AS received
									  FROM supervision_finance.ledger_allocation la
											JOIN supervision_finance.ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED'
									  WHERE la.status NOT IN ('PENDING', 'UN ALLOCATED')
										AND la.invoice_id = i.id
										AND (
											((l.type IN (SELECT ledger_type FROM receipt_transactions_types)) AND (l.created_at::DATE <= $1::DATE) AND (l.created_at IS NOT NULL))
											OR
											((l.type NOT IN (SELECT ledger_type FROM receipt_transactions_types)) AND (l.datetime::DATE <= $1::DATE))
										)
									  ) transactions ON TRUE
                                       LEFT JOIN LATERAL (
                                  SELECT ifr.supervisionlevel AS supervision_level
                                  FROM supervision_finance.invoice_fee_range ifr
                                  WHERE ifr.invoice_id = i.id
                                  ORDER BY id DESC
                                  LIMIT 1
                                  ) sl ON TRUE
							WHERE i.raiseddate <= $1::DATE AND i.amount > COALESCE(transactions.received, 0)),
     age_per_client AS (SELECT fc.client_id, MAX(oi.age) AS age
                        FROM supervision_finance.finance_client fc
                                 JOIN outstanding_invoices oi ON fc.id = oi.finance_client_id
                        GROUP BY fc.client_id)
SELECT CONCAT(p.firstname, ' ', p.surname)                 AS "Customer name",
       p.caserecnumber                                     AS "Customer number",
       fc.sop_number                                       AS "SOP number",
       d.deputytype                                        AS "Deputy type",
       COALESCE(active_orders.is_active, 'No')             AS "Active case?",
       '="0470"'                                           AS "Entity",
       '99999999'                                          AS "Receivable cost centre",
       'BALANCE SHEET'                                     AS "Receivable cost centre description",
       '1816102003'                                        AS "Receivable account code",
       cc.code                                             AS "Revenue cost centre",
       cc.cost_centre_description                          AS "Revenue cost centre description",
       a.code                                              AS "Revenue account code",
       a.account_code_description                          AS "Revenue account code description",
       oi.feetype                                          AS "Invoice type",
       oi.reference                                        AS "Trx number",
       tt.description                                      AS "Transaction description",
       TO_CHAR(oi.raiseddate, 'YYYY-MM-DD')                AS "Invoice date",
       TO_CHAR(oi.due_date, 'YYYY-MM-DD')                  AS "Due date",
       CASE
       WHEN oi.raiseddate >= DATE_TRUNC('year', oi.raiseddate) + INTERVAL '3 months'
           THEN CONCAT(EXTRACT(YEAR FROM oi.raiseddate), '/', TO_CHAR(oi.raiseddate + INTERVAL '1 year', 'YY'))
       ELSE CONCAT(EXTRACT(YEAR FROM oi.raiseddate - INTERVAL '1 year'), '/', TO_CHAR(oi.raiseddate, 'YY'))
	   END                                                 AS "Financial year",
       '30 NET'                                            AS "Payment terms",
       oi.amount                             AS "Original amount",
       oi.outstanding                        AS "Outstanding amount",
       CASE
			WHEN $1::DATE <= oi.due_date::DATE THEN oi.outstanding
			ELSE '0' END AS "Current",
		CASE
			WHEN $1::DATE > oi.due_date::DATE AND oi.age <= 1 THEN oi.outstanding
	   		ELSE '0' END AS "0-1 years",
       CASE WHEN oi.age > 1 AND oi.age <= 2 THEN oi.outstanding ELSE '0' END AS "1-2 years",
       CASE WHEN oi.age > 2 AND oi.age <= 3 THEN oi.outstanding ELSE '0' END AS "2-3 years",
       CASE WHEN oi.age > 3 AND oi.age <= 5 THEN oi.outstanding ELSE '0' END AS "3-5 years",
       CASE WHEN oi.age > 5 THEN oi.outstanding ELSE '0' END AS "5+ years",
       CASE
           WHEN apc.age < 1 THEN '="0-1"'
		   WHEN apc.age >= 1 AND apc.age < 2 THEN '="1-2"'
		   WHEN apc.age >= 2 AND apc.age < 3 THEN '="2-3"'
		   WHEN apc.age >= 3 AND apc.age < 5 THEN '="3-5"'
           ELSE '="5+"' END                                   AS "Debt impairment years"
FROM supervision_finance.finance_client fc
         JOIN outstanding_invoices oi ON fc.id = oi.finance_client_id
         JOIN age_per_client apc ON fc.client_id = apc.client_id
         JOIN supervision_finance.transaction_type tt
              ON oi.feetype = tt.fee_type AND oi.supervision_level = tt.supervision_level
         JOIN supervision_finance.account a ON tt.account_code = a.code
         JOIN supervision_finance.cost_centre cc ON cc.code = a.cost_centre
         JOIN public.persons p ON fc.client_id = p.id
         LEFT JOIN public.persons d ON p.feepayer_id = d.id
         LEFT JOIN LATERAL (
    SELECT 'Yes' AS is_active
    FROM cases c
    WHERE p.id = c.client_id
      AND c.orderstatus = 'ACTIVE'
    LIMIT 1
    ) active_orders ON TRUE;`

func (a *AgedDebt) GetHeaders() []string {
	return []string{
		"Customer name",
		"Customer number",
		"SOP number",
		"Deputy type",
		"Active case?",
		"Entity",
		"Receivable cost centre",
		"Receivable cost centre description",
		"Receivable account code",
		"Revenue cost centre",
		"Revenue cost centre description",
		"Revenue account code",
		"Revenue account code description",
		"Invoice type",
		"Trx number",
		"Transaction description",
		"Invoice date",
		"Due date",
		"Financial year",
		"Payment terms",
		"Original amount",
		"Outstanding amount",
		"Current",
		"0-1 years",
		"1-2 years",
		"2-3 years",
		"3-5 years",
		"5+ years",
		"Debt impairment years",
	}
}

func (a *AgedDebt) GetParams() []any {
	var (
		to time.Time
	)

	if a.ToDate == nil || a.ToDate.IsNull() {
		to = a.Today
	} else {
		to = a.ToDate.Time
	}

	return []any{to.Format("2006-01-02")}
}
