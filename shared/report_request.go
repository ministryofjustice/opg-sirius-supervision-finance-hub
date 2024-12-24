package shared

type ReportRequest struct {
	ReportType         string `json:"reportType"`
	ReportJournalType  string `json:"reportJournalType"`
	ReportScheduleType string `json:"reportScheduleType"`
	ReportAccountType  string `json:"reportAccountType"`
	ReportDebtType     string `json:"reportDebtType"`
	DateOfTransaction  *Date  `json:"dateOfTransaction,omitempty"`
	ToDateField        *Date  `json:"toDateField,omitempty"`
	FromDateField      *Date  `json:"fromDateField,omitempty"`
	Email              string `json:"email"`
}

func NewReportRequest(reportType, reportJournalType, reportScheduleType, reportAccountType, reportDebtType, dateOfTransaction, dateTo, dateFrom, email string) ReportRequest {
	download := ReportRequest{
		ReportType:         reportType,
		ReportJournalType:  reportJournalType,
		ReportScheduleType: reportScheduleType,
		ReportAccountType:  reportAccountType,
		ReportDebtType:     reportDebtType,
		Email:              email,
	}

	if dateOfTransaction != "" {
		raisedDateFormatted := NewDate(dateOfTransaction)
		download.DateOfTransaction = &raisedDateFormatted
	}

	if dateTo != "" {
		startDateFormatted := NewDate(dateTo)
		download.ToDateField = &startDateFormatted
	}

	if dateFrom != "" {
		endDateFormatted := NewDate(dateFrom)
		download.FromDateField = &endDateFormatted
	}

	return download
}
