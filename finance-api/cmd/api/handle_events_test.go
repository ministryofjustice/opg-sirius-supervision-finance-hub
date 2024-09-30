package api

import (
	"bytes"
	"encoding/json"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/opg-sirius-finance-hub/apierror"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
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
			name: "upload event",
			event: shared.Event{
				Source:     "aws.s3",
				DetailType: "AWS API Call via CloudTrail",
				Detail:     shared.FinanceAdminUploadEvent{RequestParameters: shared.RequestParameters{BucketName: "bucket1", Key: "file.csv"}},
			},
			expectedErr:     nil,
			expectedHandler: "ProcessFinanceAdminUpload",
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
			server := Server{Service: mock}

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
