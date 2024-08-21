package api

import (
	"bytes"
	"encoding/json"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_handleEvents(t *testing.T) {
	tests := []struct {
		name            string
		event           shared.Event
		expectedStatus  int
		expectedHandler string
	}{
		{
			name: "reapply event",
			event: shared.Event{
				Source:     "opg.supervision.sirius",
				DetailType: "debt-position-changed",
				Detail:     shared.DebtPositionChangedEvent{ClientID: 1},
			},
			expectedStatus:  http.StatusOK,
			expectedHandler: "ReapplyCredit",
		},
		{
			name: "unknown event",
			event: shared.Event{
				Source:     "opg.supervision.sirius",
				DetailType: "test",
			},
			expectedStatus:  http.StatusUnprocessableEntity,
			expectedHandler: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mock := &mockService{}
			server := Server{Service: mock}

			var body bytes.Buffer
			_ = json.NewEncoder(&body).Encode(test.event)
			r := httptest.NewRequest(http.MethodPost, "/events", &body)
			ctx := telemetry.ContextWithLogger(r.Context(), telemetry.NewLogger("test"))
			r = r.WithContext(ctx)
			w := httptest.NewRecorder()

			server.handleEvents(w, r)
			assert.Equal(t, test.expectedStatus, w.Result().StatusCode)
			assert.Equal(t, test.expectedHandler, mock.lastCalled)
		})
	}
}
