package api

import (
	"github.com/jackc/pgx/v5"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServer_getFeeReductions(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/fee-reductions", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)

	feeReductionInfo := &shared.FeeReductions{
		shared.FeeReduction{
			Id:           1,
			Type:         "REMISSION",
			StartDate:    shared.Date{Time: date},
			EndDate:      shared.Date{Time: date},
			DateReceived: shared.Date{Time: date},
			Status:       "Active",
			Notes:        "Remission notes and its active",
		},
	}

	mock := &mockService{feeReductions: feeReductionInfo}
	server := Server{Service: mock}
	server.getFeeReductions(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := `[{"id":1,"type":"REMISSION","startDate":"16\/03\/2020","endDate":"16\/03\/2020","dateReceived":"16\/03\/2020","status":"Active","notes":"Remission notes and its active"}]`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, 1, mock.expectedId)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getFeeReductions_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/fee-reductions", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrTooManyRows}
	server := Server{Service: mock}
	server.getFeeReductions(w, req)

	res := w.Result()

	assert.Equal(t, 500, res.StatusCode)
}
