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

func TestGetInvoiceAdjustmentsCanReturn200(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", ""})

	json := `
	[
	  {
		 "id":3,
		 "invoiceRef":"N2000001/20",
		 "raisedDate":"01/04/2222",
		 "adjustmentType":"CREDIT MEMO",
		 "amount":232,
		 "status":"PENDING",
		 "notes":"Some notes here"
	  }
	]
	`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(rq *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := shared.InvoiceAdjustments{
		{
			Id:             3,
			InvoiceRef:     "N2000001/20",
			RaisedDate:     shared.NewDate("01/04/2222"),
			AdjustmentType: shared.AdjustmentTypeCreditMemo,
			Amount:         232,
			Status:         "PENDING",
			Notes:          "Some notes here",
		},
	}

	resp, err := client.GetInvoiceAdjustments(testContext(), 3)

	assert.Equal(t, nil, err)
	assert.Equal(t, expectedResponse, resp)
}

func TestGetInvoiceAdjustmentsCanThrow500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	_, err := client.GetInvoiceAdjustments(testContext(), 1)

	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/invoice-adjustments",
		Method: http.MethodGet,
	}, err)
}

func TestGetInvoiceAdjustmentsUnauthorised(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	resp, err := client.GetInvoiceAdjustments(testContext(), 3)

	var expectedResponse shared.InvoiceAdjustments

	assert.Equal(t, expectedResponse, resp)
	assert.Equal(t, ErrUnauthorized, err)
}
