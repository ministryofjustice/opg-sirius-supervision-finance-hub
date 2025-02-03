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
	"os"
	"testing"
	"time"
)

func TestRequestReport(t *testing.T) {
	var b bytes.Buffer

	ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("test"))

	downloadForm := &shared.ReportRequest{
		ReportType:        "AccountsReceivable",
		ReportAccountType: "AgedDebt",
		Email:             "joseph@test.com",
	}

	_ = json.NewEncoder(&b).Encode(downloadForm)

	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/downloads", &b)
	w := httptest.NewRecorder()

	mock := &MockReports{}
	mock.requestedReport = downloadForm
	server := NewServer(nil, mock, "", nil, nil)
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
		ReportType:        "AccountsReceivable",
		ReportAccountType: "AgedDebt",
		Email:             "",
	}

	_ = json.NewEncoder(&b).Encode(downloadForm)
	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/downloads", &b)
	w := httptest.NewRecorder()

	mock := &MockReports{}
	server := NewServer(nil, mock, "", nil, nil)
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

	ctx := telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("test"))

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
				ReportType:        "Journal",
				ReportAccountType: "AgedDebt",
				Email:             "email",
				DateOfTransaction: &tt.date,
			}

			_ = json.NewEncoder(&b).Encode(downloadForm)
			r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/downloads", &b)
			w := httptest.NewRecorder()

			mock := &MockReports{}
			server := NewServer(nil, mock, "", nil, nil)
			err := server.requestReport(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.err, err)
		})
	}
}
