package server

import (
	"net/http"
)

// this might not be necessary if we are handling form submissions dynamically
type AppVars struct {
	Path            string
	XSRFToken       string
	EnvironmentVars EnvironmentVars
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
