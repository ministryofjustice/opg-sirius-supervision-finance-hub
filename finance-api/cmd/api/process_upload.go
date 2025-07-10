package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"io"
	"log/slog"
	"net/http"
	"slices"
)

type Upload struct {
	UploadType   shared.ReportUploadType
	EmailAddress string
	UploadDate   shared.Date
	PisNumber    int
	FileBytes    io.Reader
	Filename     string
}

func (s *Server) processUpload(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var upload shared.Upload
	defer unchecked(r.Body.Close)

	if err := json.NewDecoder(r.Body).Decode(&upload); err != nil {
		return apierror.BadRequestError("event", "unable to parse upload", err)
	}

	fileBytes, err := base64.StdEncoding.DecodeString(upload.Base64Data)
	if err != nil {
		return apierror.BadRequestError("event", "Invalid file data", err)
	}

	logger := s.Logger(ctx)

	logger.Info(fmt.Sprintf("processing %s upload", upload.UploadType))

	go func(logger *slog.Logger) {
		ctx := telemetry.ContextWithLogger(context.Background(), logger)
		ctx = auth.Context{
			Context: ctx,
			User:    r.Context().(auth.Context).User,
		}
		s.processUploadFile(ctx, Upload{
			UploadType:   upload.UploadType,
			EmailAddress: upload.EmailAddress,
			UploadDate:   upload.UploadDate,
			PisNumber:    upload.PisNumber,
			FileBytes:    bytes.NewReader(fileBytes),
			Filename:     upload.Filename,
		})
	}(telemetry.LoggerFromContext(ctx))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil
}

func (s *Server) processUploadFile(ctx context.Context, upload Upload) {
	var payload notify.Payload
	var err error
	logger := s.Logger(ctx)

	if upload.UploadType.IsDirectUpload() {
		err := s.service.ProcessDirectUploadReport(ctx, upload.Filename, upload.FileBytes, upload.UploadType)
		if err != nil {
			logger.Error("unable to upload report due to error", "err", err)
		}
		payload = createUploadNotifyPayload(upload.EmailAddress, upload.UploadType, err, nil)
	} else {
		csvReader := csv.NewReader(upload.FileBytes)

		records, err := csvReader.ReadAll()
		if err != nil {
			logger.Error("unable to read report", "err", err)
			payload := createUploadNotifyPayload(upload.EmailAddress, upload.UploadType, fmt.Errorf("unable to read report"), map[int]string{})
			err = s.notify.Send(ctx, payload)
			if err != nil {
				logger.Error("unable to send notification", "err", err)
			}
			return
		}

		if upload.UploadType.IsPayment() {
			failedLines, perr := s.service.ProcessPayments(ctx, records, upload.UploadType, upload.UploadDate, upload.PisNumber)
			if perr != nil {
				logger.Error("unable to process payments due to error", "err", perr)
			} else if len(failedLines) > 0 {
				logger.Error(fmt.Sprintf("unable to process payments due to %d failed lines", len(failedLines)))
			}
			payload = createUploadNotifyPayload(upload.EmailAddress, upload.UploadType, perr, failedLines)
		} else if upload.UploadType.IsReversal() {
			failedLines, perr := s.service.ProcessPaymentReversals(ctx, records, upload.UploadType)
			if perr != nil {
				logger.Error("unable to process payment reversals due to error", "err", perr)
			} else if len(failedLines) > 0 {
				logger.Error(fmt.Sprintf("unable to process payment reversals due to %d failed lines", len(failedLines)))
			}
			payload = createUploadNotifyPayload(upload.EmailAddress, upload.UploadType, err, failedLines)
		} else if upload.UploadType.IsRefund() {
			failedLines, perr := s.service.ProcessFulfilledRefunds(ctx, records, upload.UploadDate)
			if perr != nil {
				logger.Error("unable to process fulfilled refunds due to error", "err", perr)
			} else if len(failedLines) > 0 {
				logger.Error(fmt.Sprintf("unable to process fulfilled refunds due to %d failed lines", len(failedLines)))
			}
			payload = createUploadNotifyPayload(upload.EmailAddress, upload.UploadType, err, failedLines)
		} else {
			logger.Error("invalid upload type", "type", upload.UploadType)
			payload = createUploadNotifyPayload(upload.EmailAddress, upload.UploadType, fmt.Errorf("invalid upload type"), map[int]string{})
		}
	}

	err = s.notify.Send(ctx, payload)
	if err != nil {
		logger.Error("unable to send notification", "err", err)
	}
}

func formatFailedLines(failedLines map[int]string) []string {
	var errorMessage string
	var formattedLines []string
	var keys []int
	for i := range failedLines {
		keys = append(keys, i)
	}

	slices.Sort(keys)

	for _, key := range keys {
		failedLine := failedLines[key]
		errorMessage = ""

		switch failedLine {
		case validation.UploadErrorDateParse:
			errorMessage = "Unable to parse date - please use the format DD/MM/YYYY"
		case validation.UploadErrorDateTimeParse:
			errorMessage = "Unable to parse date - please use the format YYYY-MM-DD HH:MM:SS"
		case validation.UploadErrorAmountParse:
			errorMessage = "Unable to parse amount - please use the format 320.00"
		case validation.UploadErrorClientNotFound:
			errorMessage = "Could not find a client with this court reference"
		case validation.UploadErrorPaymentTypeParse:
			errorMessage = "Unable to parse payment type"
		case validation.UploadErrorUnknownUploadType:
			errorMessage = "Unknown upload type"
		case validation.UploadErrorNoMatchedPayment:
			errorMessage = "Unable to find a matched payment to reverse"
		case validation.UploadErrorReversalClientNotFound:
			errorMessage = "Could not find client with this court reference [New (correct) court reference]"
		case validation.UploadErrorDuplicateReversal:
			errorMessage = "This payment has already been reversed"
		case validation.UploadErrorRefundNotFound:
			errorMessage = "The refund could not be found - either the data does not match or the refund has been cancelled"
		case validation.UploadErrorMaximumDebt:
			errorMessage = "Payment could not be reversed - maximum invoice debt exceeded"
		}

		formattedLines = append(formattedLines, fmt.Sprintf("Line %d: %s", key, errorMessage))
	}

	return formattedLines
}

func createUploadNotifyPayload(email string, uploadType shared.ReportUploadType, err error, failedLines map[int]string) notify.Payload {
	var payload notify.Payload

	if err != nil {
		payload = notify.Payload{
			EmailAddress: email,
			TemplateId:   notify.ProcessingErrorTemplateId,
			Personalisation: struct {
				Error      string `json:"error"`
				UploadType string `json:"upload_type"`
			}{
				Error:      err.Error(),
				UploadType: uploadType.Translation(),
			},
		}
	} else if len(failedLines) != 0 {
		payload = notify.Payload{
			EmailAddress: email,
			TemplateId:   notify.ProcessingFailedTemplateId,
			Personalisation: struct {
				FailedLines []string `json:"failed_lines"`
				UploadType  string   `json:"upload_type"`
			}{
				FailedLines: formatFailedLines(failedLines),
				UploadType:  uploadType.Translation(),
			},
		}
	} else {
		payload = notify.Payload{
			EmailAddress: email,
			TemplateId:   notify.ProcessingSuccessTemplateId,
			Personalisation: struct {
				UploadType string `json:"upload_type"`
			}{uploadType.Translation()},
		}
	}

	return payload
}
