package api

import (
	"bytes"
	"github.com/opg-sirius-finance-hub/shared"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFeeReductions(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "http://localhost:8181", logger)

	json := `[
        {
            "id": 1,
            "type": "EXEMPTION",
            "startDate": "2022-04-01T00:00:00+00:00",
            "endDate": "2021-03-31T00:00:00+00:00",
            "dateReceived": "2021-02-02T00:00:00+00:00",
            "notes": "Exemption cancelled due to incorrect filing",
            "deleted": true
        },
        {
            "id": 2,
            "type": "REMISSION",
            "startDate": "2022-04-01T00:00:00+00:00",
            "endDate": "2021-03-31T00:00:00+00:00",
            "dateReceived": "2021-06-02T00:00:00+00:00",
            "notes": "Remission for 2021/2022",
            "deleted": false
        }
	]`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := shared.FeeReductions{
		{
			Id:           1,
			Type:         "EXEMPTION",
			StartDate:    shared.NewDate("2022-04-01T00:00:00+00:00"),
			EndDate:      shared.NewDate("2021-03-31T00:00:00+00:00"),
			DateReceived: shared.NewDate("2021-02-02T00:00:00+00:00"),
			Notes:        "Exemption cancelled due to incorrect filing",
			Deleted:      true,
		},
		{
			Id:           2,
			Type:         "REMISSION",
			StartDate:    shared.NewDate("2022-04-01T00:00:00+00:00"),
			EndDate:      shared.NewDate("2021-03-31T00:00:00+00:00"),
			DateReceived: shared.NewDate("2021-06-02T00:00:00+00:00"),
			Notes:        "Remission for 2021/2022",
			Deleted:      false,
		},
	}

	feeReductions, err := client.GetFeeReductions(getContext(nil), 1)
	assert.Equal(t, expectedResponse, feeReductions)
	assert.Equal(t, nil, err)
}

func TestGetFeeReductionsReturnsUnauthorisedClientError(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)
	_, err := client.GetFeeReductions(getContext(nil), 1)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestFeeReductionsReturns500Error(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	_, err := client.GetFeeReductions(getContext(nil), 1)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/api/v1/finance/1/finance-discounts",
		Method: http.MethodGet,
	}, err)
}
