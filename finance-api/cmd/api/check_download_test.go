package api

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckDownload(t *testing.T) {
	req := httptest.NewRequest(http.MethodHead, "/download?uid=eyJLZXkiOiJ0ZXN0LmNzdiIsIlZlcnNpb25JZCI6InZwckF4c1l0TFZzYjVQOUhfcUhlTlVpVTlNQm5QTmN6In0=", nil)
	w := httptest.NewRecorder()

	mockS3 := mockFileStorage{}
	mockS3.exists = true

	server := NewServer(nil, nil, &mockS3, nil, nil, nil, &Envs{ReportsBucket: "test"})
	err := server.checkDownload(w, req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCheckDownload_noMatch(t *testing.T) {
	req := httptest.NewRequest(http.MethodHead, "/download?uid=eyJLZXkiOiJ0ZXN0LmNzdiIsIlZlcnNpb25JZCI6InZwckF4c1l0TFZzYjVQOUhfcUhlTlVpVTlNQm5QTmN6In0=", nil)
	w := httptest.NewRecorder()

	mockS3 := mockFileStorage{}
	mockS3.exists = false

	server := NewServer(nil, nil, &mockS3, nil, nil, nil, &Envs{ReportsBucket: "test"})
	err := server.checkDownload(w, req)

	assert.ErrorIs(t, err, apierror.NotFound{})
}
