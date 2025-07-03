package api

import (
	"github.com/jackc/pgx/v5"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
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
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getAccountInformation(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

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
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	err := server.getAccountInformation(w, req)

	expected := apierror.NotFoundError(pgx.ErrNoRows)
	assert.ErrorAs(t, err, &expected)
}

func TestServer_getAccountInformation_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrTooManyRows}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	err := server.getAccountInformation(w, req)

	assert.Error(t, err)
}
