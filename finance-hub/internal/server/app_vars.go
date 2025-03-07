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
}

func NewAppVars(r *http.Request, envVars Envs) AppVars {
	ctx := r.Context()

	clientId := r.PathValue("clientId")
	tabs := []Tab{
		{
			Id:       "invoices",
			Title:    "Invoices",
			BasePath: "/clients/" + clientId + "/invoices",
		},
		{
			Id:       "fee-reductions",
			Title:    "Fee Reductions",
			BasePath: "/clients/" + clientId + "/fee-reductions",
		},
		{
			Id:       "pending-invoice-adjustments",
			Title:    "Pending Adjustments",
			BasePath: "/clients/" + clientId + "/pending-invoice-adjustments",
		},
		{
			Id:       "billing-history",
			Title:    "Billing History",
			BasePath: "/clients/" + clientId + "/billing-history",
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
			}
		}
	}
}
