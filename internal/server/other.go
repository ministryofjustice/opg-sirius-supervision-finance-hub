package server

import (
	"net/http"
)

type OtherTab struct {
	Error         string
	HoldingString string
}

func other(client FinanceHubInformation, tmpl Template) Handler {
	return func(app FinanceVars, w http.ResponseWriter, r *http.Request) error {
		var vars OtherTab

		vars.HoldingString = "Hello other tab!"

		return tmpl.Execute(w, vars)
	}
}
