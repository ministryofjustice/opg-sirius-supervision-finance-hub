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
	ctx := r.Context()
	clientID := getClientID(r)
	financeClient, err := h.Client().GetAccountInformation(ctx, clientID)
	if err != nil {
		return err
	}

	data := &PaymentMethodVars{financeClient.PaymentMethod, strconv.Itoa(clientID), v}
	return h.execute(w, r, data)
}
