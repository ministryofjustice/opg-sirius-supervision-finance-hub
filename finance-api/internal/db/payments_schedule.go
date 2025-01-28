package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type PaymentsSchedule struct {
	Date         shared.Date
	ScheduleType shared.ReportScheduleType
}

const PaymentsScheduleQuery = ``

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
	default:
		transactionType = shared.TransactionTypeUnknown
	}
	return []any{p.Date.Time.Format("2006-01-02 15:04:05"), transactionType.Key()} // TODO: Add golive
}
