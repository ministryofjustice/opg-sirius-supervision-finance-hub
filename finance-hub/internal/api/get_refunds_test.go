package api

import (
	"bytes"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRefundsCanReturn200(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", ""})

	json := `
		{
			"refunds": [
		  	{
				"id": 3,
				"raisedDate": "01/04/2222",
				"fulfilledDate": {
				  "value": "02/05/2222",
				  "valid": true
				},
				"amount": 232,
				"status": "PENDING",
				"notes": "Some notes here",
				"createdBy": 99,
				"bankDetails": {
				  "value": {
					"name": "Billy Banker",
					"account": "12345678",
					"sortCode": "10-20-30"
				  },
				  "valid": true
				}
			  }
			],
			"creditBalance": 50
		}
	`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(rq *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	fulfilledDate := "02/05/2222"

	expectedResponse := shared.Refunds{
		CreditBalance: 50,
		Refunds: []shared.Refund{
			{
				ID:            3,
				RaisedDate:    shared.NewDate("01/04/2222"),
				FulfilledDate: shared.TransformNillableDate(&fulfilledDate),
				Amount:        232,
				Status:        shared.RefundStatusPending,
				Notes:         "Some notes here",
				CreatedBy:     99,
				BankDetails: shared.NewNillable(
					&shared.BankDetails{
						Name:     "Billy Banker",
						Account:  "12345678",
						SortCode: "10-20-30",
					},
				),
			},
		},
	}

	resp, err := client.GetRefunds(testContext(), 3)

	assert.Equal(t, nil, err)
	assert.Equal(t, expectedResponse, resp)
}

func TestGetRefundsCanThrow500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	_, err := client.GetRefunds(testContext(), 1)

	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/refunds",
		Method: http.MethodGet,
	}, err)
}

func TestGetRefundsUnauthorised(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	resp, err := client.GetRefunds(testContext(), 3)

	var expectedResponse shared.Refunds

	assert.Equal(t, expectedResponse, resp)
	assert.Equal(t, ErrUnauthorized, err)
}
