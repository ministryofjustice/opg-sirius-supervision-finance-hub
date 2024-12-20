package api

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"unicode"
)

const s3Directory = "finance-admin"

func validateCSVHeaders(file []byte, reportUploadType shared.ReportUploadType) error {
	fileReader := bytes.NewReader(file)
	csvReader := csv.NewReader(fileReader)
	expectedHeaders := reportUploadType.CSVHeaders()

	readHeaders, err := csvReader.Read()
	if err != nil {
		return apierror.ValidationError{Errors: apierror.ValidationErrors{
			"FileUpload": {
				"read-failed": "Failed to read CSV headers",
			},
		},
		}
	}

	for i, header := range readHeaders {
		readHeaders[i] = cleanString(header)
	}

	// Compare the extracted headers with the expected headers
	if !reflect.DeepEqual(readHeaders, expectedHeaders) {
		return apierror.ValidationError{Errors: apierror.ValidationErrors{
			"FileUpload": {
				"incorrect-headers": "CSV headers do not match for the report trying to be uploaded",
			},
		},
		}
	}

	_, err = fileReader.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	return nil
}

func reportHeadersByType(reportType string) []string {
	switch reportType {
	case shared.ReportTypeUploadDeputySchedule.Key():
		return []string{"Deputy number", "Deputy name", "Case number", "Client forename", "Client surname", "Do not invoice", "Total outstanding"}
	case shared.ReportTypeUploadDebtChase.Key():
		return []string{"Client_no", "Deputy_name", "Total_debt"}
	case shared.ReportTypeUploadPaymentsOPGBACS.Key():
		return []string{"Line", "Type", "Code", "Number", "Transaction", "Value Date", "Amount", "Amount Reconciled", "Charges", "Status", "Desc Flex", "Consolidated line"}
	default:
		return []string{"Unknown report type"}
	}
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

func (s *Server) upload(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var upload shared.Upload
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&upload); err != nil {
		return err
	}

	err := validateCSVHeaders(upload.File, upload.ReportUploadType)
	if err != nil {
		return err
	}

	// TODO: update filestorage
	_, err = s.filestorage.PutFile(
		ctx,
		os.Getenv("ASYNC_S3_BUCKET"),
		fmt.Sprintf("%s/%s", s3Directory, upload.Filename),
		bytes.NewReader(upload.File))

	if err != nil {
		return err
	}

	uploadEvent := event.FinanceAdminUpload{
		EmailAddress: upload.Email,
		Filename:     fmt.Sprintf("%s/%s", s3Directory, upload.Filename),
		UploadType:   upload.ReportUploadType.Key(),
		UploadDate:   upload.UploadDate,
	}
	// TODO: Change this to send to notify
	err = s.dispatch.FinanceAdminUpload(ctx, uploadEvent)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	return nil
}
