package server

import (
	"net/http"
	"strconv"
)

type AddRefundForm struct {
	ClientId string
	AppVars
}

type AddRefundHandler struct {
	router
}

func (h *AddRefundHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	clientID := getClientID(r)

	data := AddRefundForm{ClientId: strconv.Itoa(clientID), AppVars: v}

	return h.execute(w, r, data)
}
