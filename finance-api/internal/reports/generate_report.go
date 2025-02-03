package reports

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/db"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"time"
)

const reportRequestedTemplateId = "bade69e4-0eb1-4896-a709-bd8f8371a629"

func getAccountsReceivableQuery(accountType shared.ReportAccountType, fromDate *shared.Date, toDate *shared.Date) (db.ReportQuery, error) {
	switch accountType {
	case shared.ReportAccountTypeAgedDebt:
		return &db.AgedDebt{
			FromDate: fromDate,
			ToDate:   toDate,
		}, nil
	case shared.ReportAccountTypeAgedDebtByCustomer:
		return &db.AgedDebtByCustomer{}, nil
	case shared.ReportAccountTypeARPaidInvoiceReport:
		return &db.PaidInvoices{
			FromDate: fromDate,
			ToDate:   toDate,
		}, nil
	case shared.ReportAccountTypeInvoiceAdjustments:
		return &db.InvoiceAdjustments{
			FromDate: fromDate,
			ToDate:   toDate,
		}, nil
	case shared.ReportAccountTypeBadDebtWriteOffReport:
		return &db.BadDebtWriteOff{
			FromDate: fromDate,
			ToDate:   toDate,
		}, nil
	case shared.ReportAccountTypeTotalReceiptsReport:
		return &db.Receipts{
			FromDate: fromDate,
			ToDate:   toDate,
		}, nil
	case shared.ReportAccountTypeUnappliedReceipts:
		return &db.CustomerCredit{}, nil
	}

	return nil, fmt.Errorf("Unrecognised report account type")
}

func getJournalQuery(journalType shared.ReportJournalType, date *shared.Date) (db.ReportQuery, error) {
	switch journalType {
	case shared.ReportTypeNonReceiptTransactions:
		return &db.NonReceiptTransactions{Date: date}, nil
	case shared.ReportTypeReceiptTransactions:
		return &db.ReceiptTransactions{Date: date}, nil
	}

	return nil, fmt.Errorf("Unrecognised report journal type")
}

func getQuery(reportType shared.ReportsType, reportRequest shared.ReportRequest, requestedDate time.Time) (db.ReportQuery, string, string, error) {
	switch reportType {
	case shared.ReportsTypeAccountsReceivable:
		accountType := shared.ParseReportAccountType(reportRequest.ReportAccountType)
		filename := fmt.Sprintf("%s_%s.csv", accountType.Key(), requestedDate.Format("02:01:2006"))
		friendlyName := accountType.Translation()
		query, err := getAccountsReceivableQuery(accountType, reportRequest.FromDateField, reportRequest.ToDateField)
		return query, filename, friendlyName, err
	case shared.ReportsTypeJournal:
		journalType := shared.ParseReportJournalType(reportRequest.ReportJournalType)
		filename := fmt.Sprintf("%s_%s.csv", journalType.Key(), requestedDate.Format("02:01:2006"))
		friendlyName := journalType.Translation()
		query, err := getJournalQuery(journalType, reportRequest.DateOfTransaction)
		return query, filename, friendlyName, err
	}
	return nil, "", "", fmt.Errorf("Unknown report type")
}

func (c *Client) GenerateAndUploadReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) error {
	reportType := shared.ParseReportsType(reportRequest.ReportType)

	query, filename, friendlyName, err := getQuery(reportType, reportRequest, requestedDate)
	if err != nil {
		return err
	}

	file, err := c.generate(ctx, filename, query)
	if err != nil {
		return err
	}

	defer file.Close()

	versionId, err := c.fileStorage.PutFile(
		ctx,
		c.envs.ReportsBucket,
		filename,
		file,
	)

	if err != nil {
		return err
	}

	payload, err := c.createDownloadNotifyPayload(reportRequest.Email, filename, versionId, requestedDate, friendlyName)
	if err != nil {
		return err
	}

	err = c.notify.Send(ctx, payload)
	if err != nil {
		return err
	}

	return nil
}

type reportRequestedNotifyPersonalisation struct {
	FileLink          string `json:"file_link"`
	ReportName        string `json:"report_name"`
	RequestedDate     string `json:"requested_date"`
	RequestedDateTime string `json:"requested_date_time"`
}

func (c *Client) createDownloadNotifyPayload(emailAddress string, filename string, versionId *string, requestedDate time.Time, reportName string) (notify.Payload, error) {
	if versionId == nil {
		return notify.Payload{}, fmt.Errorf("s3 version ID not found")
	}

	downloadRequest := shared.DownloadRequest{
		Key:       filename,
		VersionId: *versionId,
	}

	uid, err := downloadRequest.Encode()
	if err != nil {
		return notify.Payload{}, err
	}

	downloadLink := fmt.Sprintf("%s/download?uid=%s", c.envs.FinanceAdminURL, uid)

	payload := notify.Payload{
		EmailAddress: emailAddress,
		TemplateId:   reportRequestedTemplateId,
		Personalisation: reportRequestedNotifyPersonalisation{
			downloadLink,
			reportName,
			requestedDate.Format("2006-01-02"),
			requestedDate.Format("2006-01-02 15:04:05"),
		},
	}

	return payload, nil
}
