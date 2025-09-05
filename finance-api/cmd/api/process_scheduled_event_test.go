package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func Test_processScheduledEvent(t *testing.T) {
	tests := []struct {
		name                 string
		event                shared.ScheduledEvent
		expectedResponse     error
		expectedFunctionCall string
		expectedParams       []interface{}
		hasError             bool
	}{
		{
			name: "Unknown report",
			event: shared.ScheduledEvent{
				Trigger: "unknown",
			},
			expectedResponse: fmt.Errorf("invalid scheduled event trigger: unknown"),
			hasError:         true,
		},
		{
			name: "Negative Invoices",
			event: shared.ScheduledEvent{
				Trigger: "refund-expiry",
			},
			expectedResponse:     nil,
			hasError:             false,
			expectedFunctionCall: "ExpireRefunds",
		},
		{
			name: "Direct Debit Collection",
			event: shared.ScheduledEvent{
				Trigger: "direct-debit-collection",
			},
			expectedResponse:     nil,
			hasError:             false,
			expectedFunctionCall: "AddCollectedPayments",
			expectedParams:       []interface{}{time.Now().UTC().Truncate(24 * time.Hour)},
		},
		{
			name: "Direct Debit Collection with override",
			event: shared.ScheduledEvent{
				Trigger: "direct-debit-collection",
				Override: shared.ScheduledDirectDebitCollectionOverride{
					Date: shared.NewDate("2022-04-02"),
				},
			},
			expectedResponse:     nil,
			hasError:             false,
			expectedFunctionCall: "AddCollectedPayments",
			expectedParams:       []interface{}{time.Date(2022, 4, 2, 0, 0, 0, 0, time.UTC)},
		},
	}
	for _, tt := range tests {
		ctx := auth.Context{
			Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
			User:    &shared.User{ID: 1},
		}

		service := &mockService{}
		server := NewServer(service, nil, nil, nil, nil, nil, nil)
		
		err := server.processScheduledEvent(ctx, tt.event)
		if tt.hasError {
			assert.Error(t, err, tt.name)
		}
		assert.Equal(t, tt.expectedResponse, err, tt.name)

		if tt.expectedFunctionCall == "" {
			assert.Len(t, service.called, 0, tt.name)
		} else {
			assert.Equal(t, tt.expectedFunctionCall, service.called[0], tt.name)
			assert.Equal(t, tt.expectedParams, service.lastCalledParams, tt.name)
		}
	}
}
