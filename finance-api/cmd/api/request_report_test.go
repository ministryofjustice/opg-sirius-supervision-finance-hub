package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestRequestReport(t *testing.T) {
	var b bytes.Buffer

	ctx := auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
		User:    &shared.User{ID: 1},
	}
	subType := shared.AccountsReceivableTypeAgedDebt

	downloadForm := &shared.ReportRequest{
		ReportType:             shared.ReportsTypeAccountsReceivable,
		AccountsReceivableType: &subType,
		Email:                  "joseph@test.com",
	}

	_ = json.NewEncoder(&b).Encode(downloadForm)

	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/downloads", &b)
	w := httptest.NewRecorder()

	reports := &MockReports{}
	reports.requestedReport = downloadForm
	service := &mockService{}

	done := make(chan struct{})

	server := NewServer(service, reports, nil, nil, nil, nil, nil)

	server.onReportRequested = func() {
		close(done)
	}

	_ = server.requestReport(w, r)

	// wait for async to finish
	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for async report to complete")
	}

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, "", w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code)

	assert.EqualValues(t, downloadForm, reports.requestedReport)
	assert.Equal(t, "PostReportActions", service.lastCalled)
}

func TestRequestReportNoEmail(t *testing.T) {
	var b bytes.Buffer

	ctx := auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
		User:    &shared.User{ID: 1},
	}
	subType := shared.AccountsReceivableTypeAgedDebt

	downloadForm := &shared.ReportRequest{
		ReportType:             shared.ReportsTypeAccountsReceivable,
		AccountsReceivableType: &subType,
		Email:                  "",
	}

	_ = json.NewEncoder(&b).Encode(downloadForm)
	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/downloads", &b)
	w := httptest.NewRecorder()

	mock := &MockReports{}
	server := NewServer(nil, mock, nil, nil, nil, nil, nil)
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

func TestRequestReportJournalDate(t *testing.T) {
	os.Setenv("FINANCE_HUB_LIVE_DATE", "2024-01-01")
	var b bytes.Buffer

	ctx := auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
		User:    &shared.User{ID: 1},
	}

	tests := []struct {
		name string
		date shared.Date
		err  error
	}{
		{
			"Before live date",
			shared.NewDate("2023-01-01"),
			apierror.ValidationError{Errors: apierror.ValidationErrors{
				"Date": {
					"Date": "Date must be before today and after 2024-01-01",
				},
			},
			},
		},
		{
			"On live date",
			shared.NewDate("2024-01-01"),
			nil,
		},
		{
			"Yesterday",
			shared.NewDate(time.Now().AddDate(0, 0, -1).Format("2006-01-02")),
			nil,
		},
		{
			"Today",
			shared.NewDate(time.Now().Format("2006-01-02")),
			apierror.ValidationError{Errors: apierror.ValidationErrors{
				"Date": {
					"Date": "Date must be before today and after 2024-01-01",
				},
			},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			downloadForm := shared.ReportRequest{
				ReportType:      shared.ReportsTypeJournal,
				JournalType:     toPtr(shared.JournalTypeReceiptTransactions),
				Email:           "email",
				TransactionDate: &tt.date,
			}

			_ = json.NewEncoder(&b).Encode(downloadForm)
			r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/downloads", &b)
			w := httptest.NewRecorder()

			service := &mockService{}
			reports := &MockReports{}

			done := make(chan struct{})

			server := NewServer(service, reports, nil, nil, nil, nil, nil)
			server.onReportRequested = func() {
				close(done)
			}

			err := server.requestReport(w, r)

			if tt.err == nil {
				// wait for async to finish
				select {
				case <-done:
					// success
				case <-time.After(2 * time.Second):
					t.Fatal("timeout waiting for async report to complete")
				}
			}

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.err, err)
		})
	}
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
				ScheduleType:    toPtr(shared.ScheduleTypeOnlineCardPayments),
				TransactionDate: &shared.Date{Time: goLive},
			},
			expectedError: nil,
		},
		{
			name: "missing email",
			reportRequest: shared.ReportRequest{
				Email:           "",
				ReportType:      shared.ReportsTypeSchedule,
				ScheduleType:    toPtr(shared.ScheduleTypeOnlineCardPayments),
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
				ScheduleType:    toPtr(shared.ScheduleTypeOnlineCardPayments),
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
				ScheduleType:    toPtr(shared.ScheduleTypeOnlineCardPayments),
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
				ScheduleType:    toPtr(shared.ScheduleTypeOnlineCardPayments),
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
		{
			name: "multiple validation errors",
			reportRequest: shared.ReportRequest{
				Email:           "",
				ReportType:      shared.ReportsTypeSchedule,
				ScheduleType:    toPtr(shared.ScheduleTypeOnlineCardPayments),
				TransactionDate: &shared.Date{Time: time.Now().AddDate(0, 0, 1)},
			},
			expectedError: apierror.ValidationError{
				Errors: apierror.ValidationErrors{
					"Email": {
						"required": "This field Email needs to be looked at required",
					},
					"Date": {
						"date-in-the-past": "This field Date needs to be looked at date-in-the-past",
					},
				},
			},
		},
		{
			name: "missing accounts receivable type",
			reportRequest: shared.ReportRequest{
				Email:      "test@example.com",
				ReportType: shared.ReportsTypeAccountsReceivable,
			},
			expectedError: apierror.ValidationError{
				Errors: apierror.ValidationErrors{
					"AccountsReceivableType": {
						"required": "This field AccountsReceivableType needs to be looked at required",
					},
				},
			},
		},
		{
			name: "missing journal type",
			reportRequest: shared.ReportRequest{
				Email:      "test@example.com",
				ReportType: shared.ReportsTypeJournal,
			},
			expectedError: apierror.ValidationError{
				Errors: apierror.ValidationErrors{
					"JournalType": {
						"required": "This field JournalType needs to be looked at required",
					},
				},
			},
		},
		{
			name: "missing schedule type",
			reportRequest: shared.ReportRequest{
				Email:           "test@example.com",
				ReportType:      shared.ReportsTypeSchedule,
				TransactionDate: &shared.Date{Time: goLive},
			},
			expectedError: apierror.ValidationError{
				Errors: apierror.ValidationErrors{
					"ScheduleType": {
						"required": "This field ScheduleType needs to be looked at required",
					},
				},
			},
		},
		{
			name: "missing debt type",
			reportRequest: shared.ReportRequest{
				Email:      "test@example.com",
				ReportType: shared.ReportsTypeDebt,
			},
			expectedError: apierror.ValidationError{
				Errors: apierror.ValidationErrors{
					"DebtType": {
						"required": "This field DebtType needs to be looked at required",
					},
				},
			},
		},
		{
			name: "cheques schedule without pis number",
			reportRequest: shared.ReportRequest{
				Email:           "test@example",
				TransactionDate: &shared.Date{Time: goLive},
				ReportType:      shared.ReportsTypeSchedule,
				ScheduleType:    toPtr(shared.ScheduleTypeChequePayments),
			},
			expectedError: apierror.ValidationError{
				Errors: apierror.ValidationErrors{"PisNumber": map[string]string{"required": "This field PisNumber needs to be looked at required"}},
			},
		},
		{
			name: "cheques schedule with invalid pis number",
			reportRequest: shared.ReportRequest{
				Email:           "test@example",
				TransactionDate: &shared.Date{Time: goLive},
				ReportType:      shared.ReportsTypeSchedule,
				PisNumber:       12345,
				ScheduleType:    toPtr(shared.ScheduleTypeChequePayments),
			},
			expectedError: apierror.ValidationError{
				Errors: apierror.ValidationErrors{"PisNumber": map[string]string{"eqSix": "PIS number must be 6 digits"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{envs: &Envs{GoLiveDate: goLive}}
			err := server.validateReportRequest(tt.reportRequest)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func toPtr[T any](val T) *T {
	return &val
}
