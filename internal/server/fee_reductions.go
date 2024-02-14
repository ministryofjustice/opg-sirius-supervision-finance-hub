package server

import (
	"net/http"
)

type FeeReductionsPage struct {
	ListPage
}

func feeReductions(tmpl Template) Handler {
	return func(app FinanceVars, w http.ResponseWriter, r *http.Request) error {

		var vars FeeReductionsPage

		vars.App = app

		return tmpl.Execute(w, vars)
	}
}
