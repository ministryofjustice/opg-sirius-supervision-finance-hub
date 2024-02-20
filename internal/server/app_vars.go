package server

import (
	"net/http"
)

type AppVars struct {
	Path            string
	XSRFToken       string
	EnvironmentVars EnvironmentVars
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
	}

	return vars, nil
}
