package server

import (
	"net/http"
)

type FeeReductionsTab struct {
	AppVars
}

type FeeReductionsHandler struct {
	route
}

func (h *FeeReductionsHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	var data FeeReductionsTab
	data.AppVars = v

	h.Data = data
	return h.execute(w, r)
}
