package server

import (
	"net/http"
)

type AppVars struct {
	Path            string
	XSRFToken       string
	Tabs            []Tab
	EnvironmentVars EnvironmentVars
}

type Tab struct {
	Title    string
	BasePath string
	PushUrl  string
}

func NewAppVars(r *http.Request, envVars EnvironmentVars) (AppVars, error) {
	ctx := getContext(r)

	clientId := r.PathValue("id")
	tabs := []Tab{
		{
			Title:    "Invoices",
			BasePath: "/clients/" + clientId + "/invoices",
		},
		{
			Title:    "Fee Reductions",
			BasePath: "/clients/" + clientId + "/fee-reductions",
		},
		{
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
