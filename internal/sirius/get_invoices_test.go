package sirius

import (
	"bytes"
	"github.com/opg-sirius-finance-hub/internal/mocks"
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetInvoicesCanReturn200(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", logger)

	json := `
	[
	  {
		 "id":3,
		 "ref":"N2000001/20",
		 "status":"Unpaid",
		 "amount":"232",
		 "raisedDate":"01/04/2222",
		 "received":"22",
		 "outstandingBalance":"210",
		 "ledgers":[
			{
			   "amount":"123",
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

	mocks.GetDoFunc = func(rq *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := model.Invoices{
		{
			Id:                 3,
			Ref:                "N2000001/20",
			Status:             "Unpaid",
			Amount:             "232",
			RaisedDate:         model.NewDate("01/04/2222"),
			Received:           "22",
			OutstandingBalance: "210",
			Ledgers: []model.Ledger{
				{
					Amount:          "123",
					ReceivedDate:    model.NewDate("01/05/2222"),
					TransactionType: "Online card payment",
					Status:          "Applied",
				},
			},
			SupervisionLevels: []model.SupervisionLevel{
				{
					Level:  "General",
					Amount: "320",
					From:   model.NewDate("01/04/2019"),
					To:     model.NewDate("31/03/2020"),
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

	client, _ := NewApiClient(http.DefaultClient, svr.URL, logger)

	clientList, err := client.GetInvoices(getContext(nil), 3)

	var expectedResponse model.Invoices

	assert.Equal(t, expectedResponse, clientList)

	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/api/v1/clients/3/invoices",
		Method: http.MethodGet,
	}, err)
}
