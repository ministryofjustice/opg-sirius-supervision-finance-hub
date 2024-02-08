package server

import (
	"net/http"
)

type FinanceHubInformation interface{}

type InvoicePage struct {
	ListPage
}

func invoices(client FinanceHubInformation, tmpl Template) Handler {
	return func(app FinanceVars, w http.ResponseWriter, r *http.Request) error {

		var vars InvoicePage
		vars.App = app

		return tmpl.Execute(w, vars)
	}
}
