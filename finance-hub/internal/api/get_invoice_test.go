package api

import (
	"bytes"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetInvoiceCanReturn200(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "", logger)

	json := `
	  {
		 "id":3,
		 "ref":"N2000001/20",
		 "amount":232,
		 "raisedDate":"01/04/2222",
		 "received":12,
		 "outstandingBalance":210,
	  }
	`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(rq *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := shared.Invoice{
		Id:                 3,
		Ref:                "N2000001/20",
		Amount:             232,
		RaisedDate:         shared.NewDate("01/04/2222"),
		Received:           12,
		OutstandingBalance: 210,
	}

	invoice, err := client.GetInvoice(getContext(nil), 3, 4)

	assert.Equal(t, nil, err)
	assert.Equal(t, expectedResponse, invoice)
}

func TestGetInvoiceCanThrow500Error(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	_, err := client.GetInvoice(getContext(nil), 1, 1)

	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/invoices/1",
		Method: http.MethodGet,
	}, err)
}

func TestGetInvoiceUnauthorised(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	clientList, err := client.GetInvoice(getContext(nil), 3, 1)

	var expectedResponse shared.Invoices

	assert.Equal(t, expectedResponse, clientList)
	assert.Equal(t, ErrUnauthorized, err)
}
