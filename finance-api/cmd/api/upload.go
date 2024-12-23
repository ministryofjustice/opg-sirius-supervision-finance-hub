package api

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"io"
	"net/http"
	"reflect"
	"strings"
	"unicode"
)

type PaymentReportLine struct {
	Line             string `json:"Line"`
	Type             string `json:"Type"`
	Code             string `json:"Code"`
	Number           string `json:"Number"`
	TransactionDate  string `json:"Transaction Date"`
	ValueDate        string `json:"Value Date"`
	Amount           string `json:"Amount"`
	AmountReconciled string `json:"Amount Reconciled"`
	Charges          string `json:"Charges"`
	Status           string `json:"Status"`
	DescFlex         string `json:"Desc Flex"`
	ConsolidatedLine string `json:"Consolidated line"`
}

func headerRow() []string {
	return []string{"Line", "Type", "Code", "Number", "Transaction Date", "Value Date", "Amount", "Amount Reconciled", "Charges", "Status", "Desc Flex"}
}

func (s *Server) upload(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var upload shared.Upload
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&upload); err != nil {
		return err
	}

	report, err := readPaymentReport(upload.File)
	if err != nil {
		return err
	}

	// TODO: make async
	err = s.service.ProcessFinanceAdminUpload(ctx, report)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	return nil
}

func cleanString(s string) string {
	// Trim leading and trailing spaces
	s = strings.TrimSpace(s)
	// Remove non-printable characters
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, s)
}

func readPaymentReport(file []byte) ([]PaymentReportLine, error) {
	fileReader := bytes.NewReader(file)
	csvReader := csv.NewReader(fileReader)

	headers, err := csvReader.Read()
	if err != nil {
		return nil, apierror.ValidationError{Errors: apierror.ValidationErrors{
			"FileUpload": {
				"read-failed": "Failed to read CSV headers",
			},
		},
		}
	}

	for i, h := range headers {
		headers[i] = cleanString(h)
	}

	var (
		report []PaymentReportLine
	)

	if !reflect.DeepEqual(headers, headerRow()) {
		return nil, apierror.ValidationError{Errors: apierror.ValidationErrors{
			"FileUpload": {
				"incorrect-headers": "CSV headers do not match for the report trying to be uploaded",
			},
		},
		}
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading record:", err)
			return nil, err
		}

		data := make(map[string]string)
		for i, header := range headers {
			data[header] = record[i]
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			return nil, err
		}

		var line PaymentReportLine
		err = json.Unmarshal(jsonData, &line)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return nil, err
		}

		report = append(report, line)
	}

	return report, nil
}

func createUploadNotifyPayload(detail shared.FinanceAdminUploadProcessedEvent) notify.Payload {
	var payload notify.Payload

	uploadType := shared.ParseReportUploadType(detail.UploadType)
	if detail.Error != "" {
		payload = notify.Payload{
			EmailAddress: detail.EmailAddress,
			TemplateId:   notify.ProcessingErrorTemplateId,
			Personalisation: struct {
				Error      string `json:"error"`
				UploadType string `json:"upload_type"`
			}{
				detail.Error,
				uploadType.Translation(),
			},
		}
	} else if len(detail.FailedLines) != 0 {
		payload = notify.Payload{
			EmailAddress: detail.EmailAddress,
			TemplateId:   notify.ProcessingFailedTemplateId,
			Personalisation: struct {
				FailedLines []string `json:"failed_lines"`
				UploadType  string   `json:"upload_type"`
			}{
				formatFailedLines(detail.FailedLines),
				uploadType.Translation(),
			},
		}
	} else {
		payload = notify.Payload{
			EmailAddress: detail.EmailAddress,
			TemplateId:   notify.ProcessingSuccessTemplateId,
			Personalisation: struct {
				UploadType string `json:"upload_type"`
			}{uploadType.Translation()},
		}
	}

	return payload
}
