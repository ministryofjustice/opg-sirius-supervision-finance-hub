package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequestReport(t *testing.T) {
	var b bytes.Buffer

	ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("test"))

	downloadForm := &shared.ReportRequest{
		ReportType:             shared.ReportsTypeAccountsReceivable,
		AccountsReceivableType: shared.ReportAccountsReceivableTypeAgedDebt,
		Email:                  "joseph@test.com",
	}

	_ = json.NewEncoder(&b).Encode(downloadForm)

	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/downloads", &b)
	w := httptest.NewRecorder()

	mock := &MockReports{}
	mock.requestedReport = downloadForm
	server := NewServer(nil, mock, nil, nil, nil)
	_ = server.requestReport(w, r)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, "", w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code)

	assert.EqualValues(t, downloadForm, mock.requestedReport)
}

func TestRequestReportNoEmail(t *testing.T) {
	var b bytes.Buffer

	ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("test"))

	downloadForm := shared.ReportRequest{
		ReportType:             shared.ReportsTypeAccountsReceivable,
		AccountsReceivableType: shared.ReportAccountsReceivableTypeAgedDebt,
		Email:                  "",
	}

	_ = json.NewEncoder(&b).Encode(downloadForm)
	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/downloads", &b)
	w := httptest.NewRecorder()

	mock := &MockReports{}
	server := NewServer(nil, mock, nil, nil, nil)
	err := server.requestReport(w, r)

	res := w.Result()
	defer res.Body.Close()

	expected := apierror.ValidationError{Errors: apierror.ValidationErrors{
		"Email": {
			"required": "This field Email needs to be looked at required",
		},
	},
	}

	assert.Equal(t, expected, err)
}

func TestValidateReportRequest(t *testing.T) {
	goLive := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name          string
		reportRequest shared.ReportRequest
		expectedError error
	}{
		{
			name: "valid request",
			reportRequest: shared.ReportRequest{
				Email:           "test@example.com",
				ReportType:      shared.ReportsTypeSchedule,
				TransactionDate: &shared.Date{Time: goLive},
			},
			expectedError: nil,
		},
		{
			name: "missing email",
			reportRequest: shared.ReportRequest{
				Email:           "",
				ReportType:      shared.ReportsTypeSchedule,
				TransactionDate: &shared.Date{Time: goLive},
			},
			expectedError: apierror.ValidationError{
				Errors: apierror.ValidationErrors{
					"Email": {
						"required": "This field Email needs to be looked at required",
					},
				},
			},
		},
		{
			name: "missing transaction date for schedule report",
			reportRequest: shared.ReportRequest{
				Email:           "test@example.com",
				ReportType:      shared.ReportsTypeSchedule,
				TransactionDate: nil,
			},
			expectedError: apierror.ValidationError{
				Errors: apierror.ValidationErrors{
					"Date": {
						"required": "This field Date needs to be looked at required",
					},
				},
			},
		},
		{
			name: "transaction date in the future",
			reportRequest: shared.ReportRequest{
				Email:           "test@example.com",
				ReportType:      shared.ReportsTypeSchedule,
				TransactionDate: &shared.Date{Time: time.Now().AddDate(0, 0, 1)},
			},
			expectedError: apierror.ValidationError{
				Errors: apierror.ValidationErrors{
					"Date": {
						"date-in-the-past": "This field Date needs to be looked at date-in-the-past",
					},
				},
			},
		},
		{
			name: "transaction date before go-live date",
			reportRequest: shared.ReportRequest{
				Email:           "test@example.com",
				ReportType:      shared.ReportsTypeSchedule,
				TransactionDate: &shared.Date{Time: goLive.AddDate(0, 0, -1)},
			},
			expectedError: apierror.ValidationError{
				Errors: apierror.ValidationErrors{
					"Date": {
						"min-go-live": "This field Date needs to be looked at min-go-live",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{nil, nil, nil, nil, &Envs{GoLiveDate: goLive}}
			err := server.validateReportRequest(tt.reportRequest)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
