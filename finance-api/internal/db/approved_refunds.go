package db

// ApprovedRefunds generates a report containing all reports that have an approved status.
// This is used by the Billing team to process approved refunds by uploading them to Bankline.
// Requesting this report sets the status of all approved refunds to processing.
type ApprovedRefunds struct {
	ReportQuery
}

func NewApprovedRefunds() ReportQuery {
	return &ApprovedRefunds{
		ReportQuery: NewReportQuery(ApprovedRefundsQuery),
	}
}

const ApprovedRefundsQuery = `
	SELECT fc.court_ref                                   "Court reference",
       ((r.amount / 100.0)::NUMERIC(10, 2))::VARCHAR(255) "Amount",
       bd.name                                            "Bank account name",
       bd.account    					                  "Bank account number",
       REPLACE(bd.sort_code, '-', '')                     "Bank account sort code",
       CONCAT(ca.name, ' ', ca.surname)                   "Created by",
       CONCAT(da.name, ' ', da.surname)                   "Approved by"
		FROM supervision_finance.refund r
         	JOIN supervision_finance.finance_client fc ON fc.id = r.finance_client_id
         	JOIN supervision_finance.bank_details bd ON r.id = bd.refund_id
         	JOIN public.assignees ca ON r.created_by = ca.id
         	JOIN public.assignees da ON r.decision_by = da.id
	WHERE r.decision = 'APPROVED'
	  	AND r.processed_at IS NULL; 
`

func (a *ApprovedRefunds) GetHeaders() []string {
	return []string{
		"Court reference",
		"Amount",
		"Bank account name",
		"Bank account number",
		"Bank account sort code",
		"Created by",
		"Approved by",
	}
}

func (a *ApprovedRefunds) GetQuery() string {
	return ApprovedRefundsQuery
}

func (a *ApprovedRefunds) GetParams() []any {
	return []any{}
}
