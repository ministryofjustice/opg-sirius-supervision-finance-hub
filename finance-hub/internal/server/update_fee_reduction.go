package server

import (
	"net/http"
)

type UpdateFeeReductions struct {
	FormValues FormValues
	ClientId   string
	AppVars
}

type UpdateFeeReductionHandler struct {
	router
}

func (h *UpdateFeeReductionHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {

	data := UpdateFeeReductions{FormValues{}, r.PathValue("id"), v}

	return h.execute(w, r, data)
}
