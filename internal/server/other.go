package server

import (
	"math/rand"
	"net/http"
	"strconv"
)

type OtherTab struct {
	Error         string
	HoldingString string
}

type OtherHandler struct {
	client  *ApiClient
	tmpl    Template
	partial string
}

func (h OtherHandler) render(app PageVars, w http.ResponseWriter, r *http.Request) error {
	var vars InvoiceTab

	vars.HoldingString = "I am always 123"
	vars.PageVars = app
	return h.tmpl.Execute(w, vars)
}

func (h OtherHandler) replace(app AppVars, w http.ResponseWriter, r *http.Request) error {
	var vars InvoiceTab

	vars.HoldingString = "I am a random number: " + strconv.Itoa(rand.Int())
	return h.tmpl.ExecuteTemplate(w, h.partial, vars)
}
