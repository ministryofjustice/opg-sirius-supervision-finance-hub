package server

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/auth"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewAppVars(t *testing.T) {
	r, _ := http.NewRequest("GET", "/path", nil)
	r = r.WithContext(auth.Context{XSRFToken: "abc123"})
	r.SetPathValue("clientId", "1")

	envVars := Envs{}
	vars := NewAppVars(r, envVars)

	assert.Equal(t, AppVars{
		Path:            "/path",
		XSRFToken:       "abc123",
		EnvironmentVars: envVars,
		Tabs: []Tab{
			{
				Title:    "Invoices",
				BasePath: "/clients/1/invoices",
				Id:       "invoices",
			},
			{
				Title:    "Fee Reductions",
				BasePath: "/clients/1/fee-reductions",
				Id:       "fee-reductions",
			},
			{
				Title:    "Pending Adjustments",
				BasePath: "/clients/1/pending-invoice-adjustments",
				Id:       "pending-invoice-adjustments",
			},
			{
				Title:    "Billing History",
				BasePath: "/clients/1/billing-history",
				Id:       "billing-history",
			},
		},
	}, vars)
}
