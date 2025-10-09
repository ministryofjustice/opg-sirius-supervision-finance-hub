package shared

type ReportRequest struct {
	ReportType             ReportsType             `json:"reportType"`
	JournalType            *JournalType            `json:"journalType,omitempty"`
	ScheduleType           *ScheduleType           `json:"scheduleType,omitempty"`
	AccountsReceivableType *AccountsReceivableType `json:"AccountsReceivableType,omitempty"`
	DebtType               *DebtType               `json:"debtType,omitempty"`
	TransactionDate        *Date                   `json:"transactionDate,omitempty"`
	ToDate                 *Date                   `json:"toDate,omitempty"`
	FromDate               *Date                   `json:"fromDate,omitempty"`
	Email                  string                  `json:"email"`
	PisNumber              int                     `json:"pisNumber"`
}
