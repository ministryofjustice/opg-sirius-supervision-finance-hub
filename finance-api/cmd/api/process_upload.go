package api

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"os"
	"slices"
)

func (s *Server) processUpload(ctx context.Context, event shared.FinanceAdminUploadEvent) error {
	file, err := s.fileStorage.GetFile(ctx, os.Getenv("ASYNC_S3_BUCKET"), event.Filename)

	var payload notify.Payload

	if err != nil {
		payload = createUploadNotifyPayload(event, fmt.Errorf("unable to download report"), map[int]string{})
		return s.notify.Send(ctx, payload)
	}

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		payload = createUploadNotifyPayload(event, fmt.Errorf("unable to read report"), map[int]string{})
		return s.notify.Send(ctx, payload)
	}

	if event.UploadType.IsPayment() {
		failedLines, err := s.service.ProcessPayments(ctx, records, event.UploadType, event.UploadDate)
		payload = createUploadNotifyPayload(event, err, failedLines)
	} else if event.UploadType.IsReversal() {

	} else {
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
		case "DATE_PARSE_ERROR":
			errorMessage = "Unable to parse date - please use the format DD/MM/YYYY"
		case "DATE_TIME_PARSE_ERROR":
			errorMessage = "Unable to parse date - please use the format YYYY-MM-DD HH:MM:SS"
		case "AMOUNT_PARSE_ERROR":
			errorMessage = "Unable to parse amount - please use the format 320.00"
		case "DUPLICATE_PAYMENT":
			errorMessage = "Duplicate payment line"
		case "CLIENT_NOT_FOUND":
			errorMessage = "Could not find a client with this court reference"
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
