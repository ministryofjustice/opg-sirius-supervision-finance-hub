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
	data := &FeeReductionsTab{AppVars: v}
	data.selectTab("pending-invoice-adjustments")

	return h.execute(w, r, data)
}
