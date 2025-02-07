package server

import (
	"net/http"
	"strconv"
)

type PaymentMethodVars struct {
	PaymentMethod string
	ClientId      string
	AppVars
}

type PaymentMethodHandler struct {
	router
}

func (h *PaymentMethodHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)
	financeClient, err := h.Client().GetAccountInformation(ctx, ctx.ClientId)
	if err != nil {
		return err
	}

	data := &PaymentMethodVars{financeClient.PaymentMethod, strconv.Itoa(ctx.ClientId), v}
	return h.execute(w, r, data)
}
