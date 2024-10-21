package server

import (
	"github.com/opg-sirius-finance-hub/finance-hub/internal/api"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

type BillingHistory struct {
	User               string
	Date               shared.Date
	Event              shared.BillingEvent
	OutstandingBalance string
	CreditBalance      string
}

type BillingHistoryVars struct {
	BillingHistory []BillingHistory
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

	data := &BillingHistoryVars{h.transform(ctx, billingHistory), v}
	data.selectTab("billing-history")
	return h.execute(w, r, data)
}

func (h *BillingHistoryHandler) transform(ctx api.Context, in []shared.BillingHistory) []BillingHistory {
	var out []BillingHistory

	for _, bh := range in {
		user, _ := h.Client().GetUser(ctx, bh.User)
		out = append(out, BillingHistory{
			User:               user.DisplayName,
			Date:               bh.Date,
			Event:              bh.Event,
			OutstandingBalance: shared.IntToDecimalString(bh.OutstandingBalance),
			CreditBalance:      shared.IntToDecimalString(bh.CreditBalance),
		})
	}
	return out
}
