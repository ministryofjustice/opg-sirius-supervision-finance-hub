package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"time"
)

type AllRefunds struct {
	ReportQuery
	AllRefundsInput
}

type AllRefundsInput struct {
	FromDate *shared.Date
	ToDate   *shared.Date
}

func NewAllRefunds(input AllRefundsInput) ReportQuery {
	return &AllRefunds{
		ReportQuery:     NewReportQuery(AllRefundsQuery),
		AllRefundsInput: input,
	}
}

const AllRefundsQuery = `
	SELECT fc.court_ref                                   "Court reference",
       ((r.amount / 100.0)::NUMERIC(10, 2))::VARCHAR(255) "Amount",
       TO_CHAR(r.created_at, 'YYYY-MM-DD')                "Create date",
       CONCAT(ca.name, ' ', ca.surname)                   "Created by",
       CASE 
           WHEN da.name IS NOT NULL THEN CONCAT(da.name, ' ', da.surname) 
           ELSE '' 
           END                   						  "Approved by",
       CASE
           WHEN r.fulfilled_at IS NOT NULL THEN 'FULFILLED'
           WHEN r.cancelled_at IS NOT NULL THEN 'CANCELLED'
           WHEN r.processed_at IS NOT NULL THEN 'PROCESSING'
           ELSE r.decision
           END                                            "Status",
       CASE
           WHEN r.fulfilled_at IS NOT NULL THEN TO_CHAR(r.fulfilled_at, 'YYYY-MM-DD')
           WHEN r.cancelled_at IS NOT NULL THEN TO_CHAR(r.cancelled_at, 'YYYY-MM-DD')
           WHEN r.processed_at IS NOT NULL THEN TO_CHAR(r.processed_at, 'YYYY-MM-DD')
           WHEN r.decision_by IS NULL THEN TO_CHAR(r.created_at, 'YYYY-MM-DD')
           ELSE TO_CHAR(r.decision_at, 'YYYY-MM-DD')
           END                                            "Status Date"
	FROM supervision_finance.refund r
         JOIN supervision_finance.finance_client fc ON fc.id = r.finance_client_id
         JOIN public.assignees ca ON r.created_by = ca.id
         LEFT JOIN public.assignees da ON r.decision_by = da.id
	WHERE r.created_at::DATE BETWEEN $1 AND $2
	ORDER BY r.created_at DESC, r.id;
`

func (a *AllRefunds) GetHeaders() []string {
	return []string{
		"Court reference",
		"Amount",
		"Create date",
		"Created by",
		"Approved by",
		"Status",
		"Status Date",
	}
}

func (a *AllRefunds) GetQuery() string {
	return AllRefundsQuery
}

func (a *AllRefunds) GetParams() []any {
	var (
		from, to time.Time
	)

	if a.FromDate == nil {
		from = time.Time{}
	} else {
		from = a.FromDate.Time
	}

	if a.ToDate == nil {
		to = time.Now()
	} else {
		to = a.ToDate.Time
	}

	return []any{from.Format("2006-01-02"), to.Format("2006-01-02")}
}
