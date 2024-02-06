package server

import (
	"net/http"
)

type FinanceHubInformation interface {
}

type InvoiceTab struct {
	Error         string
	HoldingString string
	PageVars
}

type InvoicesHandler struct {
	client    *ApiClient
	tmpl      Template
	partial   string
	executeFn func() error
}

func (h InvoicesHandler) render(app PageVars, w http.ResponseWriter, r *http.Request) error {
	var vars InvoiceTab

	vars.HoldingString = "I am static!"
	vars.PageVars = app
	return h.tmpl.Execute(w, vars)
}

func (h InvoicesHandler) replace(app AppVars, w http.ResponseWriter, r *http.Request) error {
	var vars InvoiceTab

	vars.HoldingString = "I am dynamic!"
	return h.tmpl.ExecuteTemplate(w, h.partial, vars)
}
