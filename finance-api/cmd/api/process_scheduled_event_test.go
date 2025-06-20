package api

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_processScheduledEvent(t *testing.T) {
	ctx := auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
		User:    &shared.User{ID: 1},
	}

	service := &mockService{}

	server := NewServer(service, nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		event            shared.ScheduledEvent
		expectedResponse error
		hasError         bool
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
			expectedResponse: nil,
			hasError:         false,
		},
	}
	for _, tt := range tests {
		err := server.processScheduledEvent(ctx, shared.ScheduledEvent{
			Trigger: tt.event.Trigger,
		})
		if tt.hasError {
			assert.Error(t, err)
		}
		assert.Equal(t, tt.expectedResponse, err)
	}
}
