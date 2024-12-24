package db

type AgedDebtByCustomer struct{}

const AgedDebtByCustomerQuery = `WITH outstanding_invoices AS (SELECT i.id                                   AS invoice_id,
                                     i.finance_client_id,
                                     i.amount - SUM(COALESCE(la.amount, 0)) AS outstanding,
                                     CASE
                                         WHEN now()::date - i.raiseddate::date < 31 THEN 0
                                         WHEN now()::date - i.raiseddate::date < 52 THEN 1
                                         WHEN now()::date - i.raiseddate::date < 66 THEN 2
                                         WHEN now()::date - i.raiseddate::date < 96 THEN 3
                                         WHEN now()::date - i.raiseddate::date < 121 THEN 4
                                         WHEN now()::date - i.raiseddate::date < 151 THEN 5
                                         WHEN now()::date - i.raiseddate::date < 365 THEN 6
                                         WHEN now()::date - i.raiseddate::date < 761 THEN 7
                                         WHEN now()::date - i.raiseddate::date < 1126 THEN 8
                                         WHEN now()::date - i.raiseddate::date < 1826 THEN 9
                                         ELSE 10
                                         END                                AS overdue_banding
                              FROM supervision_finance.invoice i
                                       LEFT JOIN supervision_finance.ledger_allocation la ON i.id = la.invoice_id
                                  AND la.status = 'ALLOCATED'
                              GROUP BY i.id, i.amount
                              HAVING i.amount > COALESCE(SUM(la.amount), 0)),
     total_by_client AS (SELECT oi.finance_client_id,
                                (SUM(oi.outstanding) / 100.0)::NUMERIC(10, 2)::varchar(255) AS total_outstanding,
                                MAX(oi.overdue_banding)                       AS max_age
                         FROM outstanding_invoices oi
                         GROUP BY oi.finance_client_id)
SELECT CONCAT(p.firstname, ' ', p.surname)                                         "Customer Name",
       p.caserecnumber                                                             "Customer number",
       fc.sop_number                                                               "SOP number",
       d.deputytype                                                                "Deputy type",
       COALESCE(active_orders.is_active, 'No')                                     "Active case?",
       tbc.total_outstanding AS                                                    "Outstanding amount",
       CASE WHEN tbc.max_age = 0 THEN tbc.total_outstanding ELSE '0' END             "Current",
       CASE WHEN tbc.max_age = 1 THEN tbc.total_outstanding ELSE '0' END             "1-21 days",
       CASE WHEN tbc.max_age = 2 THEN tbc.total_outstanding ELSE '0' END             "22-35 days",
       CASE WHEN tbc.max_age = 3 THEN tbc.total_outstanding ELSE '0' END             "36-65 days",
       CASE WHEN tbc.max_age = 4 THEN tbc.total_outstanding ELSE '0' END             "66-90 days",
       CASE WHEN tbc.max_age = 5 THEN tbc.total_outstanding ELSE '0' END             "91-120 days",
       CASE WHEN tbc.max_age = 6 THEN tbc.total_outstanding ELSE '0' END             "121-365 days",
       CASE WHEN tbc.max_age BETWEEN 1 AND 6 THEN tbc.total_outstanding ELSE '0' END "0-1 years",
       CASE WHEN tbc.max_age = 7 THEN tbc.total_outstanding ELSE '0' END             "1-2 years",
       CASE WHEN tbc.max_age = 8 THEN tbc.total_outstanding ELSE '0' END             "2-3 years",
       CASE WHEN tbc.max_age = 9 THEN tbc.total_outstanding ELSE '0' END             "3-5 years",
       CASE WHEN tbc.max_age = 10 THEN tbc.total_outstanding ELSE '0' END            "5+ years"
FROM supervision_finance.finance_client fc
         JOIN total_by_client tbc ON fc.id = tbc.finance_client_id
         JOIN public.persons p ON fc.client_id = p.id
         LEFT JOIN public.persons d ON p.feepayer_id = d.id
         LEFT JOIN LATERAL (
    SELECT 'Yes' AS is_active
    FROM public.cases c
    WHERE p.id = c.client_id
      AND c.orderstatus = 'ACTIVE'
    LIMIT 1
    ) active_orders ON TRUE;`

func (a *AgedDebtByCustomer) GetHeaders() []string {
	return []string{
		"Customer Name",
		"Customer number",
		"SOP number",
		"Deputy type",
		"Active case?",
		"Outstanding amount",
		"Current",
		"1 - 21 Days",
		"22 - 35 Days",
		"36 - 65 Days",
		"66 - 90 Days",
		"91 - 120 Days",
		"121 - 365 Days",
		"0-1 years",
		"1-2 years",
		"2-3 years",
		"3-5 years",
		"5+ years",
	}
}

func (a *AgedDebtByCustomer) GetQuery() string {
	return AgedDebtByCustomerQuery
}

func (a *AgedDebtByCustomer) GetParams() []any {
	return []any{}
}
