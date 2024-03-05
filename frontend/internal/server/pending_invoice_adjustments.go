package server

import (
	"net/http"
)

type PendingInvoiceAdjustmentsTab struct {
	AppVars
}

type PendingInvoiceAdjustmentsHandler struct {
	router
}

func (h *PendingInvoiceAdjustmentsHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	var data FeeReductionsTab
	data.AppVars = v

	return h.execute(w, r, data)
}
