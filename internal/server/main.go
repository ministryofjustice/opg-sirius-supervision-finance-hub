package server

import (
	"net/http"
)

type ClientVars struct {
	FirstName   string
	Surname     string
	Outstanding string
}

func main(client FinanceHubInformation, tmpl Template) Handler {
	return func(app FinanceVars, w http.ResponseWriter, r *http.Request) error {
		vars := ClientVars{
			FirstName:   "Ian",
			Surname:     "Moneybags",
			Outstanding: "1000000",
		}

		return tmpl.Execute(w, vars)
	}
}
