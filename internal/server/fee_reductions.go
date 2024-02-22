package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"net/http"
)

type FeeReductionsTab struct {
	FeeReductions model.FeeReductions
	AppVars
}

type FeeReductionsHandler struct {
	router
}

func (h *FeeReductionsHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)

	feeReductions, err := h.Client().GetFeeReductions(ctx, ctx.ClientId)
	if err != nil {
		return err
	}

	data := FeeReductionsTab{feeReductions, v}
	return h.execute(w, r, data)
}
