package reports

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/db"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

const (
	reportRequestedTemplateId = "bade69e4-0eb1-4896-a709-bd8f8371a629"
	reportFailedTemplateId    = "31c40127-b5b6-4d23-aaab-050d90639d83"
)

func (c *Client) createDownloadFeeAccrualNotifyPayload(emailAddress string, requestedDate time.Time) (notify.Payload, error) {
	downloadRequest := shared.DownloadRequest{Key: "Fee_Accrual.csv"}

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
			shared.AccountsReceivableTypeFeeAccrual.Translation(),
			requestedDate.Format("2006-01-02"),
			requestedDate.Format("2006-01-02 15:04:05"),
		},
	}

	return payload, nil
}

func (c *Client) GenerateAndUploadReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) {
	logger := telemetry.LoggerFromContext(ctx)
	filename, reportName, stream, err := c.generateReport(ctx, reportRequest, requestedDate)

	if err != nil {
		logger.Error("failed to generate report", "err", err)
		notifyErr := c.sendFailureNotification(ctx, reportRequest.Email, requestedDate, reportName)
		if notifyErr != nil {
			logger.Error("unable to send message to notify", "err", notifyErr)
		}
		return
	}

	if reportRequest.ReportType == shared.ReportsTypeAccountsReceivable &&
		*reportRequest.AccountsReceivableType == shared.AccountsReceivableTypeFeeAccrual {
		payload, err := c.createDownloadFeeAccrualNotifyPayload(reportRequest.Email, requestedDate)
		if err != nil {
			logger.Error("failed to generate notify payload", "err", err)
		}

		err = c.notify.Send(ctx, payload)
		if err != nil {
			logger.Error("unable to send message to notify", "err", err)
		}
		return
	}

	versionId, err := c.fileStorage.StreamFile(ctx, c.envs.ReportsBucket, filename, stream)
	if err != nil {
		logger.Error("failed to generate report", "err", err)
		notifyErr := c.sendFailureNotification(ctx, reportRequest.Email, requestedDate, reportName)
		if notifyErr != nil {
			logger.Error("unable to send message to notify", "err", notifyErr)
		}
		return
	}

	notifyErr := c.sendSuccessNotification(ctx, reportRequest.Email, filename, versionId, requestedDate, reportName)
	if notifyErr != nil {
		logger.Error("unable to send message to notify", "err", notifyErr)
	}
}

func (c *Client) generateReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) (filename string, reportName string, stream io.ReadCloser, err error) {
	var query db.ReportQuery

	switch reportRequest.ReportType {
	case shared.ReportsTypeAccountsReceivable:
		reportDate := requestedDate.Format("02:01:2006")
		switch *reportRequest.AccountsReceivableType {
		case shared.AccountsReceivableTypeAgedDebt:
			reportDate = reportRequest.ToDate.Time.Format("02:01:2006")
			query = db.NewAgedDebt(db.AgedDebtInput{
				ToDate: reportRequest.ToDate,
			})
		case shared.AccountsReceivableTypeAgedDebtByCustomer:
			query = db.NewAgedDebtByCustomer()
		case shared.AccountsReceivableTypeARPaidInvoice:
			query = db.NewPaidInvoices(db.PaidInvoicesInput{
				FromDate:   reportRequest.FromDate,
				ToDate:     reportRequest.ToDate,
				GoLiveDate: c.envs.GoLiveDate,
			})
		case shared.AccountsReceivableTypeInvoiceAdjustments:
			query = db.NewInvoiceAdjustments(db.InvoiceAdjustmentsInput{
				FromDate:   reportRequest.FromDate,
				ToDate:     reportRequest.ToDate,
				GoLiveDate: c.envs.GoLiveDate,
			})
		case shared.AccountsReceivableTypeBadDebtWriteOff:
			query = db.NewBadDebtWriteOff(db.BadDebtWriteOffInput{
				FromDate:   reportRequest.FromDate,
				ToDate:     reportRequest.ToDate,
				GoLiveDate: c.envs.GoLiveDate,
			})
		case shared.AccountsReceivableTypeTotalReceipts:
			query = db.NewReceipts(db.ReceiptsInput{
				FromDate: reportRequest.FromDate,
				ToDate:   reportRequest.ToDate,
			})
		case shared.AccountsReceivableTypeUnappliedReceipts:
			query = db.NewCustomerCredit()
		case shared.AccountsReceivableTypeFeeAccrual:
			return filename, reportName, nil, nil
		default:
			return "", reportName, nil, fmt.Errorf("unimplemented accounts receivable query: %s", reportRequest.AccountsReceivableType.Key())
		}
		filename = fmt.Sprintf("%s_%s.csv", reportRequest.AccountsReceivableType.Key(), reportDate)
		reportName = reportRequest.AccountsReceivableType.Translation()
	case shared.ReportsTypeJournal:
		filename = fmt.Sprintf("%s_%s.csv", reportRequest.JournalType.Key(), reportRequest.TransactionDate.Time.Format("02:01:2006"))
		reportName = reportRequest.JournalType.Translation()
		switch *reportRequest.JournalType {
		case shared.JournalTypeNonReceiptTransactions:
			query = db.NewNonReceiptTransactions(db.NonReceiptTransactionsInput{Date: reportRequest.TransactionDate})
		case shared.JournalTypeReceiptTransactions:
			query = db.NewReceiptTransactions(db.ReceiptTransactionsInput{Date: reportRequest.TransactionDate})
		case shared.JournalTypeUnappliedTransactions:
			query = db.NewUnappliedTransactions(db.UnappliedTransactionsInput{Date: reportRequest.TransactionDate})
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
			shared.ScheduleTypeDirectDebitPayments,
			shared.ScheduleTypeChequePayments:
			query = db.NewPaymentsSchedule(db.PaymentsScheduleInput{
				Date:         reportRequest.TransactionDate,
				ScheduleType: reportRequest.ScheduleType,
				PisNumber:    reportRequest.PisNumber,
			})
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
			query = db.NewInvoicesSchedule(db.InvoicesScheduleInput{
				Date:         reportRequest.TransactionDate,
				ScheduleType: reportRequest.ScheduleType,
			})

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
			shared.ScheduleTypeGTWriteOffReversals,
			shared.ScheduleTypeADFeeReductionReversals,
			shared.ScheduleTypeGeneralFeeReductionReversals,
			shared.ScheduleTypeMinimalFeeReductionReversals,
			shared.ScheduleTypeGAFeeReductionReversals,
			shared.ScheduleTypeGSFeeReductionReversals,
			shared.ScheduleTypeGTFeeReductionReversals:
			query = db.NewAdjustmentsSchedule(db.AdjustmentsScheduleInput{
				Date:         reportRequest.TransactionDate,
				ScheduleType: reportRequest.ScheduleType,
			})

		case shared.ScheduleTypeUnappliedPayments,
			shared.ScheduleTypeReappliedPayments:
			query = db.NewUnapplyReapplySchedule(db.UnapplyReapplyScheduleInput{
				Date:         reportRequest.TransactionDate,
				ScheduleType: reportRequest.ScheduleType,
			})
		case shared.ScheduleTypeRefunds:
			query = db.NewRefundsSchedule(db.RefundsScheduleInput{
				Date: reportRequest.TransactionDate,
			})
		default:
			return "", reportName, nil, fmt.Errorf("unimplemented schedule query: %s", reportRequest.ScheduleType.Key())
		}
	case shared.ReportsTypeDebt:
		filename = fmt.Sprintf("debt_%s_%s.csv", reportRequest.DebtType.Key(), requestedDate.Format("02:01:2006"))
		reportName = reportRequest.DebtType.Translation()
		switch *reportRequest.DebtType {
		case shared.DebtTypeFeeChase:
			query = db.NewFeeChase()
		case shared.DebtTypeFinalFee:
			query = db.NewFinalFeeDebt()
		case shared.DebtTypeApprovedRefunds:
			query = db.NewApprovedRefunds()
		case shared.DebtTypeAllRefunds:
			query = db.NewAllRefunds(db.AllRefundsInput{
				FromDate: reportRequest.FromDate,
				ToDate:   reportRequest.ToDate,
			})
		default:
			return "", reportName, nil, fmt.Errorf("unimplemented debt query: %s", reportRequest.DebtType.Key())
		}
	default:
		return "", "unknown query", nil, fmt.Errorf("unknown query")
	}

	stream, err = c.stream(ctx, query)
	if err != nil {
		return filename, reportName, nil, err
	}

	return filename, reportName, stream, nil
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
