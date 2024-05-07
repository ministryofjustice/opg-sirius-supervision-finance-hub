package api

import (
	"github.com/jackc/pgx/v5"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_getAccountInformation(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	accountInfo := &shared.AccountInformation{
		OutstandingBalance: 12300,
		CreditBalance:      123,
		PaymentMethod:      "DEMANDED",
	}

	mock := &mockService{accountInfo: accountInfo}
	server := Server{Service: mock}
	server.getAccountInformation(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := `{"outstandingBalance":12300,"creditBalance":123,"paymentMethod":"DEMANDED"}`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, 1, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getAccountInformation_clientNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrNoRows}
	server := Server{Service: mock}
	server.getAccountInformation(w, req)

	res := w.Result()

	assert.Equal(t, 404, res.StatusCode)
}

func TestServer_getAccountInformation_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrTooManyRows}
	server := Server{Service: mock}
	server.getAccountInformation(w, req)

	res := w.Result()

	assert.Equal(t, 500, res.StatusCode)
}
