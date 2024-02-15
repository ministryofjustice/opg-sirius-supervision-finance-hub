package server

import (
	"net/http"
)

type AppVars struct {
	Path            string
	XSRFToken       string
	EnvironmentVars EnvironmentVars
	Tabs            []Tab
}

type Tab struct {
	Title    string
	BasePath string
}

func NewAppVars(r *http.Request, envVars EnvironmentVars) (AppVars, error) {
	ctx := getContext(r)

	vars := AppVars{
		Path:            r.URL.Path,
		XSRFToken:       ctx.XSRFToken,
		EnvironmentVars: envVars,
		Tabs: []Tab{
			{
				Title:    "Invoices",
				BasePath: "invoices",
			},
		},
	}

	return vars, nil
}
