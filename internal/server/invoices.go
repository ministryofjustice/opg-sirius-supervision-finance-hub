package server

import (
	"net/http"
)

type InvoicePage struct {
	ListPage
}

func invoices(tmpl Template) Handler {
	return func(app FinanceVars, w http.ResponseWriter, r *http.Request) error {

		var vars InvoicePage

		vars.App = app

		return tmpl.Execute(w, vars)
	}
}
