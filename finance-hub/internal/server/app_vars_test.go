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
				Show:     true,
			},
			{
				Title:    "Fee Reductions",
				BasePath: "/clients/1/fee-reductions",
				Id:       "fee-reductions",
				Show:     true,
			},
			{
				Title:    "Invoice Adjustments",
				BasePath: "/clients/1/invoice-adjustments",
				Id:       "invoice-adjustments",
				Show:     true,
			},
			{
				Title:    "Refunds",
				BasePath: "/clients/1/refunds",
				Id:       "refunds",
				Show:     false,
			},
			{
				Title:    "Refunds",
				BasePath: "/clients/1/refunds",
				Id:       "refunds",
			},
			{
				Title:    "Billing History",
				BasePath: "/clients/1/billing-history",
				Id:       "billing-history",
				Show:     true,
			},
		},
	}, vars)
}

func TestNewAppVars_showRefunds(t *testing.T) {
	r, _ := http.NewRequest("GET", "/path", nil)
	r = r.WithContext(auth.Context{XSRFToken: "abc123"})
	r.SetPathValue("clientId", "1")

	envVars := Envs{ShowRefunds: true}
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
				Show:     true,
			},
			{
				Title:    "Fee Reductions",
				BasePath: "/clients/1/fee-reductions",
				Id:       "fee-reductions",
				Show:     true,
			},
			{
				Title:    "Invoice Adjustments",
				BasePath: "/clients/1/invoice-adjustments",
				Id:       "invoice-adjustments",
				Show:     true,
			},
			{
				Title:    "Refunds",
				BasePath: "/clients/1/refunds",
				Id:       "refunds",
				Show:     true,
			},
			{
				Title:    "Billing History",
				BasePath: "/clients/1/billing-history",
				Id:       "billing-history",
				Show:     true,
			},
		},
	}, vars)
}
