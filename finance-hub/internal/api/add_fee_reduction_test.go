package api

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddFeeReduction(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "", logger)

	json := `{
			"id":                "1",
			"feeType":           "remission",
			"startYear":         "2025",
			"lengthOfAward":     "3",
			"dateReceived":      "15/02/2024",
			"feeReductionNotes": "Fee remission note for one award",
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Body:       r,
		}, nil
	}

	err := client.AddFeeReduction(getContext(nil), 1, "remission", "2025", "3", "15/02/2024", "Fee remission note for one award")
	assert.Equal(t, nil, err)
}

//func TestAddFeeReductionUnauthorised(t *testing.T) {
//	logger, _ := SetUpTest()
//	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusUnauthorized)
//	}))
//	defer svr.Close()
//
//	client, _ := NewApiClient(http.DefaultClient, svr.URL, "/api/v1/fee-reductions", logger)
//
//	err := client.AddFeeReduction(getContext(nil), 1, "remission", "2025", "3", "15/02/2024", "Fee remission note for one award")
//
//	assert.Equal(t, ErrUnauthorized, err.Error())
//}
//
//func TestFeeReductionReturns500Error(t *testing.T) {
//	logger, _ := SetUpTest()
//	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusInternalServerError)
//	}))
//	defer svr.Close()
//
//	client, _ := NewApiClient(http.DefaultClient, svr.URL, "", logger)
//
//	err := client.AddFeeReduction(getContext(nil), 1, "remission", "2025", "3", "15/02/2024", "Fee remission note for one award")
//	assert.Equal(t, StatusError{
//		Code:   http.StatusInternalServerError,
//		URL:    svr.URL + "/api/v1/invoices/4/ledger-entries",
//		Method: http.MethodPost,
//	}, err)
//}
