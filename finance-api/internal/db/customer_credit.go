package db

import (
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

// CustomerCredit generates a report of all customers that are in credit as of a specified date. Credit is calculated as
// the sum of all unapplied and reapplied ledger allocations up to the given date.
// If the date is not provided, it defaults to the current date.
type CustomerCredit struct {
	ReportQuery
	CustomerCreditInput
}

type CustomerCreditInput struct {
	ToDate *shared.Date
}

func NewCustomerCredit(input CustomerCreditInput) ReportQuery {
	return &CustomerCredit{
		ReportQuery:         NewReportQuery(CustomerCreditQuery),
		CustomerCreditInput: input,
	}
}

const CustomerCreditQuery = `SELECT CONCAT(p.firstname, ' ', p.surname)   AS "Customer Name",
								   p.caserecnumber                       AS "Customer number",
								   fc.sop_number                         AS "SOP number",
								   ((ABS(SUM(la.amount)) / 100.0)::NUMERIC(10, 2))::VARCHAR(255) AS credit
							FROM supervision_finance.finance_client fc
									 JOIN public.persons p ON fc.client_id = p.id
									 JOIN supervision_finance.ledger l ON fc.id = l.finance_client_id
									 JOIN supervision_finance.ledger_allocation la ON l.id = la.ledger_id
							WHERE la.status IN ('UNAPPLIED', 'REAPPLIED')
							AND l.datetime::DATE <= $1
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

func (c *CustomerCredit) GetParams() []any {
	var (
		to time.Time
	)

	if c.ToDate == nil {
		to = time.Now()
	} else {
		to = c.ToDate.Time
	}

	return []any{to.Format("2006-01-02")}
}
