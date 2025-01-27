package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type PaymentsSchedule struct {
	Date shared.Date
}

const PaymentsScheduleQuery = ``

func (b *PaymentsSchedule) GetHeaders() []string {
	return []string{
		"Court reference",
		"Invoice reference",
		"Amount",
		"Payment date",
		"Bank date",
		"Create date",
	}
}

func (b *PaymentsSchedule) GetQuery() string {
	return PaymentsScheduleQuery
}

func (b *PaymentsSchedule) GetParams() []any {
	return []any{b.Date.Time.Format("2006-01-02 15:04:05")}
}
