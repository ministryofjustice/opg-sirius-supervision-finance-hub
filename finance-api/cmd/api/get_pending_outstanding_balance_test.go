package api

import (
	"github.com/jackc/pgx/v5"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_getPendingOutstandingBalance(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	outstandingBalance := int32(12300)

	mock := &mockService{outstandingBalance: outstandingBalance}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	_ = server.getPendingOutstandingBalance(w, req)

	res := w.Result()
	defer unchecked(res.Body.Close)

	expected := `12300`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(w.Body.String()))
	assert.Equal(t, 1, mock.expectedIds[0])
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestServer_getPendingOutstandingBalance_clientNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrNoRows}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	err := server.getPendingOutstandingBalance(w, req)

	expected := apierror.NotFoundError(pgx.ErrNoRows)
	assert.ErrorAs(t, err, &expected)
}

func TestServer_getPendingOutstandingBalance_error(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clients/1", nil)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	mock := &mockService{err: pgx.ErrTooManyRows}
	server := NewServer(mock, nil, nil, nil, nil, nil, nil)
	err := server.getPendingOutstandingBalance(w, req)

	assert.Error(t, err)
}
