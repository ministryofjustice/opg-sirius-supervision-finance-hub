package api

import (
	"bytes"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_download(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/download?uid=eyJLZXkiOiJ0ZXN0LmNzdiIsIlZlcnNpb25JZCI6InZwckF4c1l0TFZzYjVQOUhfcUhlTlVpVTlNQm5QTmN6In0=", nil)
	ctx := telemetry.ContextWithLogger(r.Context(), telemetry.NewLogger("test"))
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	fileContent := "col1,col2,col3\n1,a,Z\n"

	mockS3 := mockFileStorage{}
	mockS3.data = bytes.NewBufferString(fileContent)

	server := NewServer(nil, nil, &mockS3, nil, nil, nil, &Envs{ReportsBucket: "test"})
	_ = server.download(w, r)

	res := w.Result()
	defer unchecked(res.Body.Close)

	assert.Equal(t, fileContent, w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, res.Header.Get("Content-Type"), "text/csv")
	assert.Equal(t, res.Header.Get("Content-Disposition"), "attachment; filename=test.csv")
}

func TestServer_download_noMatch(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/download?uid=eyJLZXkiOiJ0ZXN0LmNzdiIsIlZlcnNpb25JZCI6InZwckF4c1l0TFZzYjVQOUhfcUhlTlVpVTlNQm5QTmN6In0=", nil)
	ctx := telemetry.ContextWithLogger(r.Context(), telemetry.NewLogger("test"))
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	mockS3 := mockFileStorage{}
	mockS3.err = &types.NoSuchKey{}
	server := NewServer(nil, nil, &mockS3, nil, nil, nil, &Envs{ReportsBucket: "test"})

	err := server.download(w, r)

	expected := apierror.NotFoundError(&types.NoSuchKey{})
	assert.ErrorAs(t, err, &expected)
}
