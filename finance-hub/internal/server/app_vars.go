package server

import (
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

type AppVars struct {
	Path            string
	XSRFToken       string
	Tabs            []Tab
	EnvironmentVars EnvironmentVars
	Errors          shared.ValidationErrors
}

type Tab struct {
	Id       string
	Title    string
	BasePath string
	Selected bool
}

func NewAppVars(r *http.Request, envVars EnvironmentVars) (AppVars, error) {
	ctx := getContext(r)

	clientId := r.PathValue("id")
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
			Title:    "Pending Invoice Adjustments",
			BasePath: "/clients/" + clientId + "/pending-invoice-adjustments",
		},
	}

	vars := AppVars{
		Path:            r.URL.Path,
		XSRFToken:       ctx.XSRFToken,
		EnvironmentVars: envVars,
		Tabs:            tabs,
	}

	return vars, nil
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
