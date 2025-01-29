package shared

type ReportRequest struct {
	ReportType             ReportsType                  `json:"reportType"`
	JournalType            ReportJournalType            `json:"journalType"`
	ScheduleType           ReportScheduleType           `json:"scheduleType"`
	AccountsReceivableType ReportAccountsReceivableType `json:"AccountsReceivableType"`
	DebtType               ReportDebtType               `json:"debtType"`
	TransactionDate        *Date                        `json:"transactionDate,omitempty"`
	ToDate                 *Date                        `json:"toDate,omitempty"`
	FromDate               *Date                        `json:"fromDate,omitempty"`
	Email                  string                       `json:"email"`
}
