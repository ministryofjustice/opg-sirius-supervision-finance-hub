package server

import (
	"github.com/opg-sirius-supervision-finance-hub/internal/sirius"
	"net/http"
)

type FinanceVars struct {
	Path            string
	XSRFToken       string
	SuccessMessage  string
	Errors          sirius.ValidationErrors
	EnvironmentVars EnvironmentVars
}

type Tab struct {
	Title string
}

type FinanceVarsClient interface {
}

func NewFinanceVars(client FinanceVarsClient, r *http.Request, envVars EnvironmentVars) (*FinanceVars, error) {
	ctx := getContext(r)

	vars := FinanceVars{
		Path:            r.URL.Path,
		XSRFToken:       ctx.XSRFToken,
		EnvironmentVars: envVars,
	}

	return &vars, nil
}
