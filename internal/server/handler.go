package server

import (
	"io"
)

type Handler struct {
	pageVars    PageVars
	client      *ApiClient
	tmpl        Template
	partial     string
	handlerFunc func(w io.Writer, data any) error
}

func NewPageHandler(pageVars PageVars, client *ApiClient, tmpl Template) *Handler {
	h := Handler{pageVars: pageVars, client: client, tmpl: tmpl}
	hf := func(w io.Writer, data any) error {
		return h.tmpl.Execute(w, data)
	}
	h.handlerFunc = hf
	return &h
}

func NewPartialHandler(pageVars PageVars, client *ApiClient, tmpl Template, partial string) *Handler {
	h := Handler{pageVars: pageVars, client: client, tmpl: tmpl, partial: partial}
	hf := func(w io.Writer, data any) error {
		return h.tmpl.ExecuteTemplate(w, partial, data)
	}
	h.handlerFunc = hf
	return &h
}
