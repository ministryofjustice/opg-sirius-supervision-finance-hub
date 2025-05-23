package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type ChequePaymentsSchedule struct {
	Date      *shared.Date
	PisNumber int
}

const ChequePaymentsScheduleQuery = `SELECT
	fc.court_ref AS "Court reference",
	COALESCE(i.reference, '') AS "Invoice reference",
	(ABS(la.amount) / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS "Amount",
	TO_CHAR(l.datetime, 'YYYY-MM-DD') AS "Payment date",
	TO_CHAR(l.bankdate, 'YYYY-MM-DD') AS "Bank date",
	TO_CHAR(l.created_at, 'YYYY-MM-DD') AS "Create date"
	FROM supervision_finance.ledger_allocation la
	JOIN supervision_finance.ledger l ON la.ledger_id = l.id
	LEFT JOIN supervision_finance.invoice i ON i.id = la.invoice_id
	JOIN supervision_finance.finance_client fc ON fc.id = l.finance_client_id
	WHERE l.bankdate = $1 
    AND l.type = $2
	AND l.pis_number = $3;
`

func (p *ChequePaymentsSchedule) GetHeaders() []string {
	return []string{
		"Court reference",
		"Invoice reference",
		"Amount",
		"Payment date",
		"Bank date",
		"Create date",
	}
}

func (p *ChequePaymentsSchedule) GetQuery() string {
	return ChequePaymentsScheduleQuery
}

func (p *ChequePaymentsSchedule) GetParams() []any {
	return []any{p.Date.Time.Format("2006-01-02"), shared.TransactionTypeSupervisionChequePayment.Key(), p.PisNumber}
}
