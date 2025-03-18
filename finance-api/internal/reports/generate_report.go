package reports

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/db"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"os"
	"time"
)

const (
	reportRequestedTemplateId = "bade69e4-0eb1-4896-a709-bd8f8371a629"
	reportFailedTemplateId    = "31c40127-b5b6-4d23-aaab-050d90639d83"
)

func (c *Client) GenerateAndUploadReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) {
	logger := telemetry.LoggerFromContext(ctx)
	filename, reportName, file, err := c.generateReport(ctx, reportRequest, requestedDate)
	if err != nil {
		logger.Error("failed to generate report", "error", err)
		err = c.sendFailureNotification(ctx, reportRequest.Email, requestedDate, reportName)
		if err != nil {
			logger.Error("unable to send message to notify", "error", err)
		}
		return
	}

	versionId, err := c.uploadReport(ctx, filename, file)
	if err != nil {
		logger.Error("failed to generate report", "error", err)
		err = c.sendFailureNotification(ctx, reportRequest.Email, requestedDate, reportName)
		if err != nil {
			logger.Error("unable to send message to notify", "error", err)
		}
	}

	err = c.sendSuccessNotification(ctx, reportRequest.Email, filename, versionId, requestedDate, reportName)
	if err != nil {
		logger.Error("unable to send message to notify", "error", err)
	}
}

func (c *Client) generateReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) (filename string, reportName string, file *os.File, err error) {
	var query db.ReportQuery

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
			return "", reportName, nil, fmt.Errorf("unimplemented accounts receivable query: %s", reportRequest.AccountsReceivableType.Key())
		}

	case shared.ReportsTypeJournal:
		filename = fmt.Sprintf("%s_%s.csv", reportRequest.JournalType.Key(), reportRequest.TransactionDate.Time.Format("02:01:2006"))
		reportName = reportRequest.JournalType.Translation()
		switch *reportRequest.JournalType {
		case shared.JournalTypeNonReceiptTransactions:
			query = &db.NonReceiptTransactions{Date: reportRequest.TransactionDate}
		case shared.JournalTypeReceiptTransactions:
			query = &db.ReceiptTransactions{Date: reportRequest.TransactionDate}
		default:
			return "", reportName, nil, fmt.Errorf("unimplemented journal query: %s", reportRequest.JournalType.Key())
		}
	case shared.ReportsTypeSchedule:
		filename = fmt.Sprintf("schedule_%s_%s.csv", reportRequest.ScheduleType.Key(), reportRequest.TransactionDate.Time.Format("02:01:2006"))
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
			shared.ScheduleTypeSOFeeInvoicesMinimal,
			shared.ScheduleTypeGAFeeInvoices,
			shared.ScheduleTypeGSFeeInvoices,
			shared.ScheduleTypeGTFeeInvoices:
			query = &db.InvoicesSchedule{
				Date:         reportRequest.TransactionDate,
				ScheduleType: reportRequest.ScheduleType,
			}
		case shared.ScheduleTypeADFeeReductions,
			shared.ScheduleTypeGeneralFeeReductions,
			shared.ScheduleTypeMinimalFeeReductions,
			shared.ScheduleTypeGAFeeReductions,
			shared.ScheduleTypeGSFeeReductions,
			shared.ScheduleTypeGTFeeReductions,
			shared.ScheduleTypeADManualCredits,
			shared.ScheduleTypeGeneralManualCredits,
			shared.ScheduleTypeMinimalManualCredits,
			shared.ScheduleTypeGAManualCredits,
			shared.ScheduleTypeGSManualCredits,
			shared.ScheduleTypeGTManualCredits,
			shared.ScheduleTypeADManualDebits,
			shared.ScheduleTypeGeneralManualDebits,
			shared.ScheduleTypeMinimalManualDebits,
			shared.ScheduleTypeGAManualDebits,
			shared.ScheduleTypeGSManualDebits,
			shared.ScheduleTypeGTManualDebits,
			shared.ScheduleTypeADWriteOffs,
			shared.ScheduleTypeGeneralWriteOffs,
			shared.ScheduleTypeMinimalWriteOffs,
			shared.ScheduleTypeGAWriteOffs,
			shared.ScheduleTypeGSWriteOffs,
			shared.ScheduleTypeGTWriteOffs,
			shared.ScheduleTypeADWriteOffReversals,
			shared.ScheduleTypeGeneralWriteOffReversals,
			shared.ScheduleTypeMinimalWriteOffReversals,
			shared.ScheduleTypeGAWriteOffReversals,
			shared.ScheduleTypeGSWriteOffReversals,
			shared.ScheduleTypeGTWriteOffReversals:
			query = &db.AdjustmentsSchedule{
				Date:         reportRequest.TransactionDate,
				ScheduleType: reportRequest.ScheduleType,
			}
		case shared.ScheduleTypeUnappliedPayments,
			shared.ScheduleTypeReappliedPayments:
			query = &db.UnapplyReapplySchedule{
				Date:         reportRequest.TransactionDate,
				ScheduleType: reportRequest.ScheduleType,
			}
		default:
			return "", reportName, nil, fmt.Errorf("unimplemented schedule query: %s", reportRequest.ScheduleType.Key())
		}
	default:
		return "", "unknown query", nil, fmt.Errorf("unknown query")
	}

	file, err = c.generate(ctx, filename, query)
	if err != nil {
		return filename, reportName, nil, err
	}
	defer file.Close()

	return filename, reportName, file, nil
}

func (c *Client) uploadReport(ctx context.Context, filename string, file *os.File) (*string, error) {
	versionId, err := c.fileStorage.PutFile(ctx, c.envs.ReportsBucket, filename, file)
	if err != nil {
		return nil, err
	}

	return versionId, nil
}

type reportRequestedNotifyPersonalisation struct {
	FileLink          string `json:"file_link"`
	ReportName        string `json:"report_name"`
	RequestedDate     string `json:"requested_date"`
	RequestedDateTime string `json:"requested_date_time"`
}

func (c *Client) sendSuccessNotification(ctx context.Context, emailAddress, filename string, versionId *string, requestedDate time.Time, reportName string) error {
	if versionId == nil {
		return fmt.Errorf("s3 version ID not found")
	}

	downloadRequest := shared.DownloadRequest{
		Key:       filename,
		VersionId: *versionId,
	}

	uid, err := downloadRequest.Encode()
	if err != nil {
		return err
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

	return c.notify.Send(ctx, payload)
}

type reportFailedNotifyPersonalisation struct {
	ReportName        string `json:"report_name"`
	RequestedDate     string `json:"requested_date"`
	RequestedDateTime string `json:"requested_date_time"`
}

func (c *Client) sendFailureNotification(ctx context.Context, emailAddress string, requestedDate time.Time, reportName string) error {
	payload := notify.Payload{
		EmailAddress: emailAddress,
		TemplateId:   reportFailedTemplateId,
		Personalisation: reportFailedNotifyPersonalisation{
			reportName,
			requestedDate.Format("2006-01-02"),
			requestedDate.Format("2006-01-02 15:04:05"),
		},
	}

	return c.notify.Send(ctx, payload)
}
