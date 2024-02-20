package server

import (
	"net/http"
)

type PendingInvoiceAdjustmentsTab struct {
	AppVars
}

type PendingInvoiceAdjustmentsHandler struct {
	route
}

func (h *PendingInvoiceAdjustmentsHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	var data FeeReductionsTab
	data.AppVars = v

	h.Data = data
	return h.execute(w, r)
}
