package server

import (
	"net/http"
	"strconv"
)

type AddRefund struct {
	ClientId string
	AppVars
}

type AddRefundHandler struct {
	router
}

func (h *AddRefundHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	clientID := getClientID(r)

	data := AddRefund{ClientId: strconv.Itoa(clientID), AppVars: v}

	return h.execute(w, r, data)
}
