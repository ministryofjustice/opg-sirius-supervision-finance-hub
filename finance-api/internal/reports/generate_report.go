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

func (c *Client) createDownloadFeeAccrualNotifyPayload(emailAddress string, requestedDate time.Time) (notify.Payload, error) {
	filename := "Fee_Accrual.csv"

	downloadRequest := shared.DownloadRequest{
		Key:    filename,                   //Create download request with no version ID to get latest
		Bucket: c.envs.LegacyReportsBucket, // Change to legacy finance bucket!
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
			"Fee Accrual",
			requestedDate.Format("2006-01-02"),
			requestedDate.Format("2006-01-02 15:04:05"),
		},
	}

	return payload, nil
}

func (c *Client) GenerateAndUploadReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) error {
	var query db.ReportQuery
	var err error

	accountType := shared.ParseReportAccountType(reportRequest.ReportAccountType)
	filename := fmt.Sprintf("%s_%s.csv", accountType.Key(), requestedDate.Format("02:01:2006"))

	switch reportRequest.ReportType {
	case "AccountsReceivable":
		switch accountType {
		case shared.ReportAccountTypeAgedDebt:
			query = &db.AgedDebt{
				FromDate: reportRequest.FromDateField,
				ToDate:   reportRequest.ToDateField,
			}
		case shared.ReportAccountTypeAgedDebtByCustomer:
			query = &db.AgedDebtByCustomer{}
		case shared.ReportAccountTypeARPaidInvoiceReport:
			query = &db.PaidInvoices{
				FromDate:   reportRequest.FromDateField,
				ToDate:     reportRequest.ToDateField,
				GoLiveDate: c.envs.GoLiveDate,
			}
		case shared.ReportAccountTypeInvoiceAdjustments:
			query = &db.InvoiceAdjustments{
				FromDate:   reportRequest.FromDateField,
				ToDate:     reportRequest.ToDateField,
				GoLiveDate: c.envs.GoLiveDate,
			}
		case shared.ReportAccountTypeBadDebtWriteOffReport:
			query = &db.BadDebtWriteOff{
				FromDate:   reportRequest.FromDateField,
				ToDate:     reportRequest.ToDateField,
				GoLiveDate: c.envs.GoLiveDate,
			}
		case shared.ReportAccountTypeTotalReceiptsReport:
			query = &db.Receipts{
				FromDate: reportRequest.FromDateField,
				ToDate:   reportRequest.ToDateField,
			}
		case shared.ReportAccountTypeUnappliedReceipts:
			query = &db.CustomerCredit{}
		case shared.ReportAccountTypeFeeAccrual:
			payload, err := c.createDownloadFeeAccrualNotifyPayload(reportRequest.Email, requestedDate)
			if err != nil {
				return err
			}

			fmt.Println(payload.Personalisation)

			return c.notify.Send(ctx, payload)
		default:
			return fmt.Errorf("unknown query")
		}
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

	payload, err := c.createDownloadNotifyPayload(reportRequest.Email, filename, versionId, requestedDate, accountType.Translation())
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
		Bucket:    c.envs.ReportsBucket,
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
