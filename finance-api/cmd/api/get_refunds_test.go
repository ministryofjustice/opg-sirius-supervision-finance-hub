package api

import (
	"github.com/jackc/pgx/v5"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_getRefunds(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/refunds", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()
	fulfilledDate := shared.NewDate("2020-04-17")

	refunds := shared.Refunds{
		CreditBalance: 50,
		Refunds: []shared.Refund{
			{
				ID:            1,
				RaisedDate:    shared.NewDate("2020-03-16"),
				FulfilledDate: shared.NewNillable(&fulfilledDate),
				Amount:        123400,
				Status:        shared.RefundStatusPending,
				Notes:         "Refund for client",
				BankDetails: shared.NewNillable(
					&shared.BankDetails{
						Name:     "Clint Client",
						Account:  "12345678",
						SortCode: "11-22-33",
					}),
				CreatedBy: 99,
			},
		},
	}

	mock := &mockService{refunds: refunds}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getRefunds(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := `{"refunds":[{"id":1,"raisedDate":"16\/03\/2020","fulfilledDate":{"Value":"17\/04\/2020","Valid":true},"amount":123400,"status":"PENDING","notes":"Refund for client","bankDetails":{"Value":{"name":"Clint Client","account":"12345678","sortCode":"11-22-33"},"Valid":true},"createdBy":99}],"creditBalance":50}`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, 1, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_get_refunds_returnsEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/2/refunds", nil)
	req.SetPathValue("clientId", "2")
	w := httptest.NewRecorder()

	refunds := shared.Refunds{}

	mock := &mockService{refunds: refunds}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getRefunds(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := `{"refunds":null,"creditBalance":0}`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, 2, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getRefunds_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1/refunds", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrTooManyRows}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	err := server.getRefunds(w, req)

	assert.Error(t, err)
}
