package api

import (
	"github.com/jackc/pgx/v5"
	"github.com/opg-sirius-finance-hub/apierror"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_getAccountInformation(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	accountInfo := &shared.AccountInformation{
		OutstandingBalance: 12300,
		CreditBalance:      123,
		PaymentMethod:      "DEMANDED",
	}

	mock := &mockService{accountInfo: accountInfo}
	server := Server{Service: mock}
	_ = server.getAccountInformation(w, req)

	res := w.Result()
	defer res.Body.Close()

	expected := `{"outstandingBalance":12300,"creditBalance":123,"paymentMethod":"DEMANDED"}`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, 1, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getAccountInformation_clientNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrNoRows}
	server := Server{Service: mock}
	err := server.getAccountInformation(w, req)

	var e apierror.NotFound
	assert.ErrorAs(t, err, &e)
}

func TestServer_getAccountInformation_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrTooManyRows}
	server := Server{Service: mock}
	err := server.getAccountInformation(w, req)

	assert.Error(t, err)
}
