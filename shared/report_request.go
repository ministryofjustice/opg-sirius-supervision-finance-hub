package shared

type ReportRequest struct {
	ReportType      ReportType `json:"reportType"`
	TransactionDate *Date      `json:"transactionDate,omitempty"`
	FromDate        *Date      `json:"fromDate,omitempty"`
	ToDate          *Date      `json:"toDate,omitempty"`
	Email           string     `json:"email"`
}
