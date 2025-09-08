package govuk

import (
	"context"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
)

func testContext() context.Context {
	return auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("test")),
	}
}
