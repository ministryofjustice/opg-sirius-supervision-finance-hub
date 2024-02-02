package server

import (
	"net/http"
)

type FinanceHubInformation interface {
}

type InvoiceTab struct {
	Error         string
	HoldingString string
	ClientVars
}

func invoices(client FinanceHubInformation, tmpl Template) Handler {
	return func(app FinanceVars, w http.ResponseWriter, r *http.Request) error {
		var vars InvoiceTab

		vars.HoldingString = "Hello invoices tab!"
		vars.ClientVars = app.ClientVars

		if r.Header.Get("HX-Request") == "true" {
			return tmpl.ExecuteTemplate(w, "invoices", vars)
		}
		return tmpl.Execute(w, vars)
	}
}
