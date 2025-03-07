package auth

import (
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestXsrfCheck(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		formToken      string
		contextToken   string
		expectedStatus int
	}{
		{
			name:           "Valid token",
			method:         http.MethodPost,
			formToken:      "valid-token",
			contextToken:   "valid-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid token",
			method:         http.MethodPost,
			formToken:      "invalid-token",
			contextToken:   "valid-token",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Missing token",
			method:         http.MethodPost,
			formToken:      "",
			contextToken:   "valid-token",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Missing token - PUT",
			method:         http.MethodPut,
			formToken:      "",
			contextToken:   "valid-token",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Missing token - DELETE",
			method:         http.MethodDelete,
			formToken:      "",
			contextToken:   "valid-token",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Skip GET",
			method:         http.MethodGet,
			formToken:      "",
			contextToken:   "valid-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Skip HEAD",
			method:         http.MethodHead,
			formToken:      "",
			contextToken:   "valid-token",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(url.Values{"xsrfToken": {tt.formToken}}.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			ctx := Context{
				Context:   telemetry.ContextWithLogger(req.Context(), telemetry.NewLogger("test")),
				XSRFToken: tt.contextToken,
			}

			req = req.WithContext(ctx)

			handler := XsrfCheck(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusForbidden {
				expectedBody := "XSRF token mismatch\n"
				if rr.Body.String() != expectedBody {
					t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
				}
			}
		})
	}
}
