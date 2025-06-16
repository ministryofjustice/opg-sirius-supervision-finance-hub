package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type UnapplyReapplySchedule struct {
	ReportQuery
	Date         *shared.Date
	ScheduleType *shared.ScheduleType
}

func NewUnapplyReapplySchedule(date *shared.Date, scheduleType *shared.ScheduleType) ReportQuery {
	return &UnapplyReapplySchedule{
		ReportQuery:  NewReportQuery(UnapplyReapplyScheduleQuery),
		Date:         date,
		ScheduleType: scheduleType,
	}
}

const UnapplyReapplyScheduleQuery = `SELECT
	fc.court_ref AS "Court reference",
	i.reference AS "Invoice reference",
	(ABS(la.amount) / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS "Amount",
	TO_CHAR(l.created_at, 'YYYY-MM-DD') AS "Created date"
	FROM supervision_finance.ledger l
	    JOIN supervision_finance.ledger_allocation la ON l.id = la.ledger_id
	    JOIN supervision_finance.finance_client fc ON fc.id = l.finance_client_id
	    JOIN supervision_finance.invoice i ON i.id = la.invoice_id
	WHERE l.created_at::DATE = $1 AND la.status = $2;
`

func (u *UnapplyReapplySchedule) GetHeaders() []string {
	return []string{
		"Court reference",
		"Invoice reference",
		"Amount",
		"Created date",
	}
}

func (u *UnapplyReapplySchedule) GetParams() []any {
	var (
		allocationStatus string
	)
	switch *u.ScheduleType {
	case shared.ScheduleTypeUnappliedPayments:
		allocationStatus = "UNAPPLIED"
	case shared.ScheduleTypeReappliedPayments:
		allocationStatus = "REAPPLIED"
	default:
		allocationStatus = ""
	}

	return []any{u.Date.Time.Format("2006-01-02"), allocationStatus}
}
