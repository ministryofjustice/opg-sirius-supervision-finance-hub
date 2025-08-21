package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/validation"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_addPendingCollection(t *testing.T) {
	var b bytes.Buffer

	data := shared.PendingCollection{
		Amount:         12345,
		CollectionDate: shared.NewDate("2019-01-27"),
	}
	_ = json.NewEncoder(&b).Encode(data)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/pending-collections", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	_ = server.addPendingCollection(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "AddPendingCollection", mock.lastCalled)
}

func TestServer_addPendingCollection500Error(t *testing.T) {
	var b bytes.Buffer

	data := shared.PendingCollection{
		Amount:         12345,
		CollectionDate: shared.NewDate("2019-01-27"),
	}
	_ = json.NewEncoder(&b).Encode(data)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/pending-collections", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{err: errors.New("something is wrong")}
	server := NewServer(mock, nil, nil, nil, nil, validator, nil)
	err := server.addPendingCollection(w, req)
	assert.Error(t, err)
}
