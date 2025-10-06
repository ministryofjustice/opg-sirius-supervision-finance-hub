package server

import (
	"context"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

type BillingHistory struct {
	User               string
	Date               shared.Date
	Event              shared.BillingEvent
	OutstandingBalance string
	CreditBalance      string
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
	logger := telemetry.LoggerFromContext(ctx)

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
			OutstandingBalance: shared.IntToDecimalString(bh.OutstandingBalance),
			CreditBalance:      shared.IntToDecimalString(bh.CreditBalance),
		})
	}

	return out
}
