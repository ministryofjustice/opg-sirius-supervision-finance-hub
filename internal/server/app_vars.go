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
	Id       string
	Title    string
	BasePath string
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
