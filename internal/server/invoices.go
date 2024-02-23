package server

import (
	"net/http"
)

type InvoiceTab struct {
	AppVars
}

type InvoicesHandler struct {
	router
}

func (h *InvoicesHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	var data InvoiceTab
	data.AppVars = v

	return h.execute(w, r, data)
}
