package api

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"os"
	"slices"
)

func (s *Server) processUpload(ctx context.Context, event shared.FinanceAdminUploadEvent) error {
	logger := s.Logger(ctx)

	logger.Info(fmt.Sprintf("processing %s upload", event.UploadType))
	file, err := s.fileStorage.GetFile(ctx, os.Getenv("ASYNC_S3_BUCKET"), event.Filename)

	var payload notify.Payload

	if err != nil {
		logger.Error("unable to fetch report from file storage", "err", err)
		payload = createUploadNotifyPayload(event, fmt.Errorf("unable to download report"), map[int]string{})
		return s.notify.Send(ctx, payload)
	}

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		logger.Error("unable to read report as CSV", "err", err)
		payload = createUploadNotifyPayload(event, fmt.Errorf("unable to read report"), map[int]string{})
		return s.notify.Send(ctx, payload)
	}

	if event.UploadType.IsPayment() {
		failedLines, perr := s.service.ProcessPayments(ctx, records, event.UploadType, event.UploadDate, event.PisNumber)
		if perr != nil {
			logger.Error("unable to process payments due to error", "err", perr)
		} else if len(failedLines) > 0 {
			logger.Error(fmt.Sprintf("unable to process payments due to %d failed lines", len(failedLines)))
		}
		payload = createUploadNotifyPayload(event, perr, failedLines)
	} else if event.UploadType.IsReversal() {
		failedLines, perr := s.service.ProcessPaymentReversals(ctx, records, event.UploadType)
		if perr != nil {
			logger.Error("unable to process payment reversals due to error", "err", perr)
		} else if len(failedLines) > 0 {
			logger.Error(fmt.Sprintf("unable to process payment reversals due to %d failed lines", len(failedLines)))
		}
		payload = createUploadNotifyPayload(event, err, failedLines)
	} else {
		logger.Error("invalid upload type", "type", event.UploadType)
		payload = createUploadNotifyPayload(event, fmt.Errorf("invalid upload type"), map[int]string{})
	}

	return s.notify.Send(ctx, payload)
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
		case validation.UploadErrorDuplicatePayment:
			errorMessage = "Duplicate payment line"
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
		}

		formattedLines = append(formattedLines, fmt.Sprintf("Line %d: %s", key, errorMessage))
	}

	return formattedLines
}

func createUploadNotifyPayload(detail shared.FinanceAdminUploadEvent, err error, failedLines map[int]string) notify.Payload {
	var payload notify.Payload

	if err != nil {
		payload = notify.Payload{
			EmailAddress: detail.EmailAddress,
			TemplateId:   notify.ProcessingErrorTemplateId,
			Personalisation: struct {
				Error      string `json:"error"`
				UploadType string `json:"upload_type"`
			}{
				Error:      err.Error(),
				UploadType: detail.UploadType.Translation(),
			},
		}
	} else if len(failedLines) != 0 {
		payload = notify.Payload{
			EmailAddress: detail.EmailAddress,
			TemplateId:   notify.ProcessingFailedTemplateId,
			Personalisation: struct {
				FailedLines []string `json:"failed_lines"`
				UploadType  string   `json:"upload_type"`
			}{
				FailedLines: formatFailedLines(failedLines),
				UploadType:  detail.UploadType.Translation(),
			},
		}
	} else {
		payload = notify.Payload{
			EmailAddress: detail.EmailAddress,
			TemplateId:   notify.ProcessingSuccessTemplateId,
			Personalisation: struct {
				UploadType string `json:"upload_type"`
			}{detail.UploadType.Translation()},
		}
	}

	return payload
}
