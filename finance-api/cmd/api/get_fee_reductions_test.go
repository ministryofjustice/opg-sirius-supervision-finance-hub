package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func TestServer_getFeeReductions(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/fee-reductions", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()
	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)

	feeReductionInfo := shared.FeeReductions{
		shared.FeeReduction{
			Id:           1,
			Type:         shared.FeeReductionTypeRemission,
			StartDate:    shared.Date{Time: date},
			EndDate:      shared.Date{Time: date},
			DateReceived: shared.Date{Time: date},
			Status:       shared.StatusActive,
			Notes:        "Remission notes and its active",
		},
	}

	mock := &mockService{feeReductions: feeReductionInfo}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getFeeReductions(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := `[{"id":1,"type":"REMISSION","startDate":"16\/03\/2020","endDate":"16\/03\/2020","dateReceived":"16\/03\/2020","status":"Active","notes":"Remission notes and its active"}]`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, 1, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getFeeReductions_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/fee-reductions", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{errs: map[string]error{"GetFeeReductions": pgx.ErrTooManyRows}}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	err := server.getFeeReductions(w, req)

	assert.Error(t, err)
}
