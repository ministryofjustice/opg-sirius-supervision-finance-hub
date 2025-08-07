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

func TestGetPermittedAdjustments(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", "http://localhost:8181"}, nil)

	json := `["CREDIT MEMO","DEBIT MEMO"]`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := []shared.AdjustmentType{shared.AdjustmentTypeCreditMemo, shared.AdjustmentTypeDebitMemo}

	types, err := client.GetPermittedAdjustments(testContext(), 1, 2)
	assert.Equal(t, expectedResponse, types)
	assert.Equal(t, nil, err)
}

func TestGetPermittedAdjustmentsReturnsUnauthorisedClientError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL}, nil)
	_, err := client.GetPermittedAdjustments(testContext(), 1, 2)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestGetPermittedAdjustmentsReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL}, nil)

	_, err := client.GetPermittedAdjustments(testContext(), 1, 2)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/invoices/2/permitted-adjustments",
		Method: http.MethodGet,
	}, err)
}
