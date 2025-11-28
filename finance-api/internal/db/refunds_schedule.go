package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

// RefundsSchedule generates a report of all refunds for a given bank date.
// This is used by the Cash Control team to reconcile refunds, and each refund schedule should correlate to a line
// in the receipts transactions journal by line description.
type RefundsSchedule struct {
	ReportQuery
	RefundsScheduleInput
}

type RefundsScheduleInput struct {
	Date *shared.Date
}

func NewRefundsSchedule(input RefundsScheduleInput) ReportQuery {
	return &RefundsSchedule{
		ReportQuery:          NewReportQuery(RefundsScheduleQuery),
		RefundsScheduleInput: input,
	}
}

const RefundsScheduleQuery = `SELECT
	fc.court_ref AS "Court reference",
	(ABS(la.amount) / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS "Amount",
	TO_CHAR(l.bankdate, 'YYYY-MM-DD') AS "Bank date",
	TO_CHAR(l.created_at, 'YYYY-MM-DD') AS "Fulfilled (create) date"
	FROM supervision_finance.ledger l
	    JOIN supervision_finance.ledger_allocation la ON l.id = la.ledger_id
	    JOIN supervision_finance.finance_client fc ON fc.id = l.finance_client_id
	WHERE l.bankdate = $1 AND l.status = 'CONFIRMED' AND la.status = 'REAPPLIED' AND la.invoice_id IS NULL;
`

func (u *RefundsSchedule) GetHeaders() []string {
	return []string{
		"Court reference",
		"Amount",
		"Bank date",
		"Fulfilled (create) date",
	}
}

func (u *RefundsSchedule) GetParams() []any {
	return []any{u.Date.Time.Format("2006-01-02")}
}
