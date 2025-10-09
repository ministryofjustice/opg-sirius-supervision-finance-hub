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

func Test_processAdhocEvent(t *testing.T) {
	ctx := auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
		User:    &shared.User{ID: 1},
	}

	service := &mockService{}

	server := NewServer(service, nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		adhocProcessName shared.AdhocEvent
		expectedResponse error
		hasError         bool
	}{
		{
			name: "Unknown report",
			adhocProcessName: shared.AdhocEvent{
				Task: "unknown",
			},
			expectedResponse: fmt.Errorf("invalid adhoc process: unknown"),
			hasError:         true,
		},
		{
			name: "Rebalance CCB",
			adhocProcessName: shared.AdhocEvent{
				Task: "RebalanceCCB",
			},
			expectedResponse: nil,
			hasError:         false,
		},
	}
	for _, tt := range tests {
		err := server.processAdhocEvent(ctx, shared.AdhocEvent{
			Task: tt.adhocProcessName.Task,
		})
		if tt.hasError {
			assert.Error(t, err)
		}
		assert.Equal(t, tt.expectedResponse, err)
	}
}
