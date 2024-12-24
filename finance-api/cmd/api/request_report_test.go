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

	mock := &mockService{}
	server := NewServer(mock, nil, nil)
	_ = server.requestReport(w, r)

	res := w.Result()
	defer res.Body.Close()

	expected := ""

	assert.Equal(t, expected, w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code)
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

	mock := &mockService{}
	server := NewServer(mock, nil, nil)
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
