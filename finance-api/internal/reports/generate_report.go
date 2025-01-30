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

func (c *Client) GenerateAndUploadReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) error {
	var (
		query      db.ReportQuery
		err        error
		filename   string
		reportName string
	)

	switch reportRequest.ReportType {
	case shared.ReportsTypeAccountsReceivable:
		filename = fmt.Sprintf("%s_%s.csv", reportRequest.AccountsReceivableType.Key(), requestedDate.Format("02:01:2006"))
		reportName = reportRequest.AccountsReceivableType.Translation()
		switch *reportRequest.AccountsReceivableType {
		case shared.AccountsReceivableTypeAgedDebt:
			query = &db.AgedDebt{
				FromDate: reportRequest.FromDate,
				ToDate:   reportRequest.ToDate,
			}
		case shared.AccountsReceivableTypeAgedDebtByCustomer:
			query = &db.AgedDebtByCustomer{}
		case shared.AccountsReceivableTypeARPaidInvoice:
			query = &db.PaidInvoices{
				FromDate:   reportRequest.FromDate,
				ToDate:     reportRequest.ToDate,
				GoLiveDate: c.envs.GoLiveDate,
			}
		case shared.AccountsReceivableTypeInvoiceAdjustments:
			query = &db.InvoiceAdjustments{
				FromDate:   reportRequest.FromDate,
				ToDate:     reportRequest.ToDate,
				GoLiveDate: c.envs.GoLiveDate,
			}
		case shared.AccountsReceivableTypeBadDebtWriteOff:
			query = &db.BadDebtWriteOff{
				FromDate:   reportRequest.FromDate,
				ToDate:     reportRequest.ToDate,
				GoLiveDate: c.envs.GoLiveDate,
			}
		case shared.AccountsReceivableTypeTotalReceipts:
			query = &db.Receipts{
				FromDate: reportRequest.FromDate,
				ToDate:   reportRequest.ToDate,
			}
		case shared.AccountsReceivableTypeUnappliedReceipts:
			query = &db.CustomerCredit{}
		default:
			return fmt.Errorf("unimplemented accounts receivable query: %s", reportRequest.AccountsReceivableType.Key())
		}
	case shared.ReportsTypeSchedule:
		filename = fmt.Sprintf("schedule_%s_%s.csv", reportRequest.ScheduleType.Key(), requestedDate.Format("02:01:2006"))
		reportName = reportRequest.ScheduleType.Translation()
		switch *reportRequest.ScheduleType {
		case shared.ScheduleTypeMOTOCardPayments,
			shared.ScheduleTypeOnlineCardPayments,
			shared.ScheduleTypeOPGBACSTransfer,
			shared.ScheduleTypeSupervisionBACSTransfer,
			shared.ScheduleTypeDirectDebitPayments:
			query = &db.PaymentsSchedule{
				Date:         reportRequest.TransactionDate,
				ScheduleType: reportRequest.ScheduleType,
			}
		case shared.ScheduleTypeAdFeeInvoices,
			shared.ScheduleTypeS2FeeInvoices,
			shared.ScheduleTypeS3FeeInvoices,
			shared.ScheduleTypeB2FeeInvoices,
			shared.ScheduleTypeB3FeeInvoices,
			shared.ScheduleTypeSFFeeInvoicesGeneral,
			shared.ScheduleTypeSFFeeInvoicesMinimal,
			shared.ScheduleTypeSEFeeInvoicesGeneral,
			shared.ScheduleTypeSEFeeInvoicesMinimal,
			shared.ScheduleTypeSOFeeInvoicesGeneral,
			shared.ScheduleTypeSOFeeInvoicesMinimal:
			query = &db.InvoicesSchedule{
				Date:         reportRequest.TransactionDate,
				ScheduleType: reportRequest.ScheduleType,
			}
		case shared.ScheduleTypeADFeeReductions,
			shared.ScheduleTypeGeneralManualCredits,
			shared.ScheduleTypeMinimalManualCredits,
			shared.ScheduleTypeADWriteOffs,
			shared.ScheduleTypeGeneralWriteOffs,
			shared.ScheduleTypeMinimalWriteOffs:
			query = &db.CreditsSchedule{
				Date:         reportRequest.TransactionDate,
				ScheduleType: reportRequest.ScheduleType,
			}
		default:
			return fmt.Errorf("unimplemented schedule query: %s", reportRequest.ScheduleType.Key())
		}
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
		c.envs.ReportsBucket,
		filename,
		file,
	)

	if err != nil {
		return err
	}

	payload, err := c.createDownloadNotifyPayload(reportRequest.Email, filename, versionId, requestedDate, reportName)
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
