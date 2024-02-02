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
	ClientVars      ClientVars
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

	myDetails := model.Assignee{}
	clientVars := ClientVars{
		FirstName:   "Ian",
		Surname:     "Moneybags",
		Outstanding: "1000000",
	}

	vars := FinanceVars{
		Path:            r.URL.Path,
		XSRFToken:       ctx.XSRFToken,
		MyDetails:       myDetails,
		ClientVars:      clientVars,
		EnvironmentVars: envVars,
	}

	return &vars, nil
}
