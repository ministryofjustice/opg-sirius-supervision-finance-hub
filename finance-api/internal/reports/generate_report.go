package reports

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/db"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"os"
	"time"
)

const reportRequestedTemplateId = "bade69e4-0eb1-4896-a709-bd8f8371a629"

func (c *Client) GenerateAndUploadReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) error {
	var query db.ReportQuery
	var err error

	filename := fmt.Sprintf("%s_%s.csv", reportRequest.ReportType.Key(), requestedDate.Format("02:01:2006"))

	switch reportRequest.ReportType {
	case shared.ReportTypeAgedDebt:
		query = &db.AgedDebt{
			FromDate: reportRequest.FromDate,
			ToDate:   reportRequest.ToDate,
		}
	case shared.ReportTypeAgedDebtByCustomer:
		query = &db.AgedDebtByCustomer{}
	case shared.ReportTypeBadDebtWriteOff:
		query = &db.BadDebtWriteOff{
			FromDate: reportRequest.FromDate,
			ToDate:   reportRequest.ToDate,
		}
	case shared.ReportTypeTotalReceipts:
		query = &db.Receipts{
			FromDate: reportRequest.FromDate,
			ToDate:   reportRequest.ToDate,
		}
	case shared.ReportTypeUnappliedReceipts:
		query = &db.CustomerCredit{}
	default:
		return fmt.Errorf("unknown query")
	}

	file, err := c.generate(ctx, filename, query)
	if err != nil {
		return err
	}

	defer file.Close()

	versionId, err := c.fileStorage.PutFile(
		ctx,
		os.Getenv("REPORTS_S3_BUCKET"),
		filename,
		file,
	)

	if err != nil {
		return err
	}

	payload, err := createDownloadNotifyPayload(reportRequest.Email, filename, versionId, requestedDate, reportRequest.ReportType.Translation())
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

func createDownloadNotifyPayload(emailAddress string, filename string, versionId *string, requestedDate time.Time, reportName string) (notify.Payload, error) {
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

	downloadLink := fmt.Sprintf("%s%s/download?uid=%s", os.Getenv("SIRIUS_PUBLIC_URL"), os.Getenv("FINANCE_ADMIN_PREFIX"), uid)

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
