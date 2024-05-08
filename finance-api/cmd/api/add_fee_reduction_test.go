package api

import (
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServer_addFeeReductions(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)

	feeReductionInfo := &shared.AddFeeReduction{
		FeeType:       "remission",
		StartYear:     "2022",
		LengthOfAward: 1,
		DateReceived:  shared.Date{Time: date},
		Notes:         "Adding a remission reduction",
	}

	mock := &mockService{feeReduction: feeReductionInfo}
	server := Server{Service: mock}
	server.addFeeReduction(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)

	expected := ""

	assert.Equal(t, expected, string(data))
	assert.Equal(t, nil, err)
}
