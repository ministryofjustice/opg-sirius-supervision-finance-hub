package server

import (
	"net/http"
)

type InvoiceTab struct {
	AppVars
}

type InvoicesHandler struct {
	route
}

func (h *InvoicesHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	var data InvoiceTab
	data.AppVars = v

	h.Data = data
	return h.execute(w, r)
}
