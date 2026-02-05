package server

import (
	"context"
	"net/http"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type BillingHistory struct {
	User               string
	Date               shared.Date
	Event              shared.BillingEvent
	OutstandingBalance int
	CreditBalance      int
}

type BillingHistoryTab struct {
	BillingHistory []BillingHistory
	AppVars
}

type BillingHistoryHandler struct {
	router
}

func (h *BillingHistoryHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	clientID := getClientID(r)

	billingHistory, err := h.Client().GetBillingHistory(ctx, clientID)
	if err != nil {
		return err
	}

	data := &BillingHistoryTab{h.transform(ctx, billingHistory), v}
	data.selectTab("billing-history")
	return h.execute(w, r, data)
}

func (h *BillingHistoryHandler) transform(ctx context.Context, in []shared.BillingHistory) []BillingHistory {
	logger := h.logger(ctx)

	var out []BillingHistory

	for _, bh := range in {
		user, err := h.Client().GetUser(ctx, bh.User)
		if err != nil {
			logger.Error("error fetching user from cache", "error", err)
		}
		out = append(out, BillingHistory{
			User:               user.DisplayName,
			Date:               bh.Date,
			Event:              bh.Event,
			OutstandingBalance: bh.OutstandingBalance,
			CreditBalance:      bh.CreditBalance,
		})
	}

	return out
}
