package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/opg-sirius-finance-hub/internal/sirius"
	"net/http"
)

type FinanceHubInformation interface {
	GetCurrentUserDetails(sirius.Context) (model.Assignee, error)
}

type InvoicePage struct {
	Error         string
	HoldingString string
}

func invoices(client FinanceHubInformation, tmpl Template) Handler {
	return func(app FinanceVars, w http.ResponseWriter, r *http.Request) error {
		ctx := getContext(r)
		myDetails, err := client.GetCurrentUserDetails(ctx)

		if err != nil {
			return err
		}

		var vars = myDetails

		return tmpl.Execute(w, vars)
	}
}
