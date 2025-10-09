package server

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

type AppVars struct {
	Path            string
	XSRFToken       string
	User            *shared.User
	Tabs            []Tab
	EnvironmentVars Envs
	Errors          apierror.ValidationErrors
	Error           string
	Code            int
}

type Tab struct {
	Id       string
	Title    string
	BasePath string
	Selected bool
	Show     bool
}

func NewAppVars(r *http.Request, envVars Envs) AppVars {
	ctx := r.Context()

	clientId := r.PathValue("clientId")
	tabs := []Tab{
		{
			Id:       "invoices",
			Title:    "Invoices",
			BasePath: "/clients/" + clientId + "/invoices",
			Show:     true,
		},
		{
			Id:       "fee-reductions",
			Title:    "Fee Reductions",
			BasePath: "/clients/" + clientId + "/fee-reductions",
			Show:     true,
		},
		{
			Id:       "invoice-adjustments",
			Title:    "Invoice Adjustments",
			BasePath: "/clients/" + clientId + "/invoice-adjustments",
			Show:     true,
		},
		{
			Id:       "refunds",
			Title:    "Refunds",
			BasePath: "/clients/" + clientId + "/refunds",
			Show:     true,
		},
		{
			Id:       "billing-history",
			Title:    "Billing History",
			BasePath: "/clients/" + clientId + "/billing-history",
			Show:     true,
		},
	}

	vars := AppVars{
		Path:            r.URL.Path,
		XSRFToken:       ctx.(auth.Context).XSRFToken,
		User:            ctx.(auth.Context).User,
		EnvironmentVars: envVars,
		Tabs:            tabs,
	}

	return vars
}

func (a *AppVars) selectTab(s string) {
	for i, tab := range a.Tabs {
		if tab.Id == s {
			a.Tabs[i] = Tab{
				tab.Id,
				tab.Title,
				tab.BasePath,
				true,
				tab.Show,
			}
		}
	}
}
