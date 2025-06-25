package api

import (
	"bytes"
	"encoding/json"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_handleEvents(t *testing.T) {
	var e apierror.BadRequest
	tests := []struct {
		name            string
		event           shared.Event
		expectedErr     error
		expectedHandler string
	}{
		{
			name: "reapply event",
			event: shared.Event{
				Source:     "opg.supervision.sirius",
				DetailType: "debt-position-changed",
				Detail:     shared.DebtPositionChangedEvent{ClientID: 1},
			},
			expectedErr:     nil,
			expectedHandler: "ReapplyCredit",
		},
		{
			name: "client created event",
			event: shared.Event{
				Source:     "opg.supervision.sirius",
				DetailType: "client-created",
				Detail:     shared.ClientCreatedEvent{ClientID: 1, CourtRef: "12345678"},
			},
			expectedErr:     nil,
			expectedHandler: "UpdateClient",
		},
		{
			name: "adhoc event",
			event: shared.Event{
				Source:     "opg.supervision.finance.adhoc",
				DetailType: "finance-adhoc",
				Detail:     shared.AdhocEvent{Task: "NegativeInvoices"},
			},
			expectedErr:     nil,
			expectedHandler: "ProcessAdhocEvent",
		},
		{
			name: "scheduled event",
			event: shared.Event{
				Source:     "opg.supervision.infra",
				DetailType: "scheduled-event",
				Detail:     shared.ScheduledEvent{Trigger: "refund-expiry"},
			},
			expectedErr:     nil,
			expectedHandler: "ExpireRefunds",
		},
		{
			name: "unknown event",
			event: shared.Event{
				Source:     "opg.supervision.sirius",
				DetailType: "test",
			},
			expectedErr:     e,
			expectedHandler: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mock := &mockService{}
			fileStorage := &mockFileStorage{}
			fileStorage.data = io.NopCloser(strings.NewReader("test"))
			notifyClient := &mockNotify{}
			server := NewServer(mock, nil, fileStorage, notifyClient, nil, nil, nil)

			var body bytes.Buffer
			_ = json.NewEncoder(&body).Encode(test.event)
			r := httptest.NewRequest(http.MethodPost, "/events", &body)
			ctx := telemetry.ContextWithLogger(r.Context(), telemetry.NewLogger("test"))
			r = r.WithContext(ctx)
			w := httptest.NewRecorder()

			err := server.handleEvents(w, r)
			if test.expectedErr != nil {
				assert.ErrorAs(t, err, &test.expectedErr)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, test.expectedHandler, mock.lastCalled)
		})
	}
}
