package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type PaymentsSchedule struct {
	Date         shared.Date
	ScheduleType shared.ReportScheduleType
}

const PaymentsScheduleQuery = `SELECT
	fc.court_ref AS "Court reference",
	i.reference AS "Invoice reference",
	(la.amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS "Amount",
	TO_CHAR(l.datetime, 'YYYY-MM-DD') AS "Payment date",
	TO_CHAR(l.bankdate, 'YYYY-MM-DD') AS "Bank date",
	TO_CHAR(l.created_at, 'YYYY-MM-DD') AS "Create date"
	FROM supervision_finance.ledger_allocation la
	LEFT JOIN supervision_finance.ledger l ON la.ledger_id = l.id
	LEFT JOIN supervision_finance.invoice i ON i.id = la.invoice_id
	JOIN supervision_finance.finance_client fc ON fc.id = l.finance_client_id
	WHERE la.status = 'ALLOCATED'
	AND l.bankdate = $1 AND l.type = $2;
`

func (p *PaymentsSchedule) GetHeaders() []string {
	return []string{
		"Court reference",
		"Invoice reference",
		"Amount",
		"Payment date",
		"Bank date",
		"Create date",
	}
}

func (p *PaymentsSchedule) GetQuery() string {
	return PaymentsScheduleQuery
}

func (p *PaymentsSchedule) GetParams() []any {
	var transactionType shared.TransactionType
	switch p.ScheduleType {
	case shared.ReportTypeMOTOCardPayments:
		transactionType = shared.TransactionTypeMotoCardPayment
	case shared.ReportTypeOnlineCardPayments:
		transactionType = shared.TransactionTypeOnlineCardPayment
	case shared.ReportOPGBACSTransfer:
		transactionType = shared.TransactionTypeOPGBACSPayment
	case shared.ReportSupervisionBACSTransfer:
		transactionType = shared.TransactionTypeSupervisionBACSPayment
	case shared.ReportDirectDebitPayments:
		transactionType = shared.TransactionTypeDirectDebitPayment
	default:
		transactionType = shared.TransactionTypeUnknown
	}
	return []any{p.Date.Time.Format("2006-01-02 15:04:05"), transactionType.Key()}
}
