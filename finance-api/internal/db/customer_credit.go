package db

import "github.com/jackc/pgx/v5"

type CustomerCredit struct{}

const CustomerCreditQuery = `SELECT CONCAT(p.firstname, ' ', p.surname)   AS "Customer Name",
								   p.caserecnumber                       AS "Customer number",
								   fc.sop_number                         AS "SOP number",
								   ((ABS(SUM(la.amount)) / 100.0)::NUMERIC(10, 2))::VARCHAR(255) AS credit
							FROM supervision_finance.finance_client fc
									 JOIN public.persons p ON fc.client_id = p.id
									 JOIN supervision_finance.ledger l ON fc.id = l.finance_client_id
									 JOIN supervision_finance.ledger_allocation la ON l.id = la.ledger_id
							WHERE la.status IN ('UNAPPLIED', 'REAPPLIED')
							GROUP BY p.caserecnumber, CONCAT(p.firstname, ' ', p.surname), fc.sop_number
							HAVING SUM(la.amount) < 0;` // #nosec G101 -- False Positive

func (c *CustomerCredit) GetHeaders() []string {
	return []string{
		"Customer name",
		"Customer number",
		"SOP number",
		"Credit balance",
	}
}

func (c *CustomerCredit) GetQuery() string {
	return CustomerCreditQuery
}

func (c *CustomerCredit) GetParams() []any {
	return []any{}
}

func (c *CustomerCredit) GetCallback() func(row pgx.CollectableRow) ([]string, error) {
	return RowToStringMap
}
