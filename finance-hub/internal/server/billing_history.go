package server

import (
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

type BillingHistoryVars struct {
	BillingHistory []shared.BillingHistory
	AppVars
}

type BillingHistoryHandler struct {
	router
}

func (h *BillingHistoryHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)

	billingHistory, err := h.Client().GetBillingHistory(ctx, ctx.ClientId)
	if err != nil {
		return err
	}

	data := &BillingHistoryVars{billingHistory, v}
	data.selectTab("billing-history")
	return h.execute(w, r, data)
}
