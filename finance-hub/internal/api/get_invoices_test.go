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

func TestGetInvoicesCanReturn200(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "", logger)

	json := `
	[
	  {
		 "id":3,
		 "ref":"N2000001/20",
		 "status":"Unpaid",
		 "amount":232,
		 "raisedDate":"01/04/2222",
		 "received":12,
		 "outstandingBalance":210,
		 "ledgers":[
			{
			   "amount":12000,
			   "receivedDate":"01/05/2222",
			   "transactionType":"Online card payment",
			   "status":"Applied"
			}
		 ],
		 "supervisionLevels":[
			{
			   "Level":"General",
			   "Amount":"320",
			   "From":"01/04/2019",
			   "To":"31/03/2020"
			}
		 ]
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

	expectedResponse := shared.Invoices{
		{
			Id:                 3,
			Ref:                "N2000001/20",
			Status:             "Unpaid",
			Amount:             232,
			RaisedDate:         shared.NewDate("01/04/2222"),
			Received:           12,
			OutstandingBalance: 210,
			Ledgers: []shared.Ledger{
				{
					Amount:          12000,
					ReceivedDate:    shared.NewDate("01/05/2222"),
					TransactionType: "Online card payment",
					Status:          "Applied",
				},
			},
			SupervisionLevels: []shared.SupervisionLevel{
				{
					Level:  "General",
					Amount: "320",
					From:   shared.NewDate("01/04/2019"),
					To:     shared.NewDate("31/03/2020"),
				},
			},
		},
	}

	invoiceList, err := client.GetInvoices(getContext(nil), 3)

	assert.Equal(t, nil, err)
	assert.Equal(t, expectedResponse, invoiceList)
}

func TestGetInvoicesCanThrow500Error(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	_, err := client.GetInvoices(getContext(nil), 1)

	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/clients/1/invoices",
		Method: http.MethodGet,
	}, err)
}

func TestGetInvoicesUnauthorised(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	clientList, err := client.GetInvoices(getContext(nil), 3)

	var expectedResponse shared.Invoices

	assert.Equal(t, expectedResponse, clientList)
	assert.Equal(t, ErrUnauthorized, err)
}
