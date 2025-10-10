package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
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
			name: "client made inactive event",
			event: shared.Event{
				Source:     "opg.supervision.sirius",
				DetailType: "client-made-inactive",
				Detail:     shared.ClientMadeInactiveEvent{ClientID: 1, CourtRef: "12345678", Surname: "Smith"},
			},
			expectedErr:     nil,
			expectedHandler: "CancelDirectDebitMandate",
		},
		{
			name: "dd invoice created event",
			event: shared.Event{
				Source:     "opg.supervision.sirius",
				DetailType: "dd-invoice-created",
				Detail:     shared.DirectDebitInvoiceCreatedEvent{ClientID: 1, CourtRef: "12345678", Surname: "Smith", InvoiceId: 1},
			},
			expectedErr:     nil,
			expectedHandler: "CreateDirectDebitScheduleForInvoice",
		},
		{
			name: "adhoc event",
			event: shared.Event{
				Source:     "opg.supervision.finance.adhoc",
				DetailType: "finance-adhoc",
				Detail:     shared.AdhocEvent{Task: "RebalanceCCB"},
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
			if test.expectedHandler != "" {
				assert.Equal(t, test.expectedHandler, mock.called[0])
			} else {
				assert.Len(t, mock.called, 0)
			}
		})
	}
}
