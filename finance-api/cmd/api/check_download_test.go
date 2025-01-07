package api

import (
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckDownload(t *testing.T) {
	r := httptest.NewRequest(http.MethodHead, "/download?uid=eyJLZXkiOiJ0ZXN0LmNzdiIsIlZlcnNpb25JZCI6InZwckF4c1l0TFZzYjVQOUhfcUhlTlVpVTlNQm5QTmN6In0=", nil)
	ctx := telemetry.ContextWithLogger(r.Context(), telemetry.NewLogger("test"))
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	mockS3 := MockFileStorage{}
	mockS3.exists = true

	server := NewServer(nil, &mockS3, nil)
	err := server.checkDownload(w, r)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCheckDownload_noMatch(t *testing.T) {
	r := httptest.NewRequest(http.MethodHead, "/download?uid=eyJLZXkiOiJ0ZXN0LmNzdiIsIlZlcnNpb25JZCI6InZwckF4c1l0TFZzYjVQOUhfcUhlTlVpVTlNQm5QTmN6In0=", nil)
	ctx := telemetry.ContextWithLogger(r.Context(), telemetry.NewLogger("test"))
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	mockS3 := MockFileStorage{}
	mockS3.exists = false

	server := NewServer(nil, &mockS3, nil)
	err := server.checkDownload(w, r)

	assert.ErrorIs(t, err, apierror.NotFound{})
}
