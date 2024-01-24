package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/opg-sirius-finance-hub/internal/sirius"
	"net/http"
)

type FinanceVars struct {
	Path            string
	XSRFToken       string
	MyDetails       model.Assignee
	SuccessMessage  string
	Errors          sirius.ValidationErrors
	EnvironmentVars EnvironmentVars
}

type Tab struct {
	Title string
}

type FinanceVarsClient interface {
	GetCurrentUserDetails(sirius.Context) (model.Assignee, error)
}

func NewFinanceVars(client FinanceVarsClient, r *http.Request, envVars EnvironmentVars) (*FinanceVars, error) {
	ctx := getContext(r)

	myDetails, err := client.GetCurrentUserDetails(ctx)
	if err != nil {
		return nil, err
	}

	vars := FinanceVars{
		Path:            r.URL.Path,
		XSRFToken:       ctx.XSRFToken,
		MyDetails:       myDetails,
		EnvironmentVars: envVars,
	}

	return &vars, nil
}
