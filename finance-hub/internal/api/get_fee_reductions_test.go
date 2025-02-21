package api

import (
	"bytes"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFeeReductions(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", ""})

	json := `[
        {
            "id": 1,
            "type": "EXEMPTION",
            "startDate": "2019-04-01T00:00:00+00:00",
            "endDate": "2021-03-31T00:00:00+00:00",
            "dateReceived": "2021-02-02T00:00:00+00:00",
			"status": "Expired",
            "notes": "Exemption cancelled due to incorrect filing"
        },
        {
            "id": 2,
            "type": "REMISSION",
            "startDate": "2019-04-01T00:00:00+00:00",
            "endDate": "2021-03-31T00:00:00+00:00",
            "dateReceived": "2021-06-02T00:00:00+00:00",
			"status": "Expired",
            "notes": "Remission for 2021/2022"
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
			Type:         shared.FeeReductionTypeExemption,
			StartDate:    shared.NewDate("2019-04-01T00:00:00+00:00"),
			EndDate:      shared.NewDate("2021-03-31T00:00:00+00:00"),
			DateReceived: shared.NewDate("2021-02-02T00:00:00+00:00"),
			Status:       "Expired",
			Notes:        "Exemption cancelled due to incorrect filing",
		},
		{
			Id:           2,
			Type:         shared.FeeReductionTypeRemission,
			StartDate:    shared.NewDate("2019-04-01T00:00:00+00:00"),
			EndDate:      shared.NewDate("2021-03-31T00:00:00+00:00"),
			DateReceived: shared.NewDate("2021-06-02T00:00:00+00:00"),
			Status:       "Expired",
			Notes:        "Remission for 2021/2022",
		},
	}

	feeReductions, err := client.GetFeeReductions(testContext(), 1)
	assert.Equal(t, expectedResponse, feeReductions)
	assert.Equal(t, nil, err)
}

func TestGetFeeReductionsReturnsUnauthorisedClientError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})
	_, err := client.GetFeeReductions(testContext(), 1)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestFeeReductionsReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	_, err := client.GetFeeReductions(testContext(), 1)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/fee-reductions",
		Method: http.MethodGet,
	}, err)
}
