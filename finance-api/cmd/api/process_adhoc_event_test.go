package api

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func Test_processAdhocEvent(t *testing.T) {
	ctx := auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
		User:    &shared.User{ID: 1},
	}

	service := &mockService{}

	server := NewServer(service, nil, nil, nil, nil, nil, nil)
	err := server.processAdhocEvent(ctx, shared.AdhocEvent{
		Task: "SomeEvent",
	})
	assert.Nil(t, err)
}

func Test_processAdhocEvent_error(t *testing.T) {
	ctx := auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
		User:    &shared.User{ID: 1},
	}

	service := &mockService{errs: map[string]error{"ProcessAdhocEvent": assert.AnError}}

	server := NewServer(service, nil, nil, nil, nil, nil, nil)
	err := server.processAdhocEvent(ctx, shared.AdhocEvent{
		Task: "SomeEvent",
	})
	assert.Error(t, err)
}
