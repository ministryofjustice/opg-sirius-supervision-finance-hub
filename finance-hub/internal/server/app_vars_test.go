package server

import (
	"github.com/opg-sirius-finance-hub/finance-hub/internal/config"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewAppVars(t *testing.T) {
	r, _ := http.NewRequest("GET", "/path", nil)
	r.SetPathValue("id", "1")

	envVars := config.EnvironmentVars{}
	vars, err := NewAppVars(r, envVars)

	assert.Nil(t, err)
	assert.Equal(t, AppVars{
		Path:            "/path",
		XSRFToken:       "",
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
				Title:    "Pending Invoice Adjustments",
				BasePath: "/clients/1/pending-invoice-adjustments",
				Id:       "pending-invoice-adjustments",
			},
		},
	}, vars)
}
