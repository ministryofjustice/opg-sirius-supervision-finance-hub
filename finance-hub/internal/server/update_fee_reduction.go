package server

import (
	"net/http"
)

type UpdateFeeReductions struct {
	FeeReductionFormValues FeeReductionFormValues
	ClientId               string
	AppVars
}

type UpdateFeeReductionHandler struct {
	router
}

func (h *UpdateFeeReductionHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {

	data := UpdateFeeReductions{FeeReductionFormValues{}, r.PathValue("id"), v}

	return h.execute(w, r, data)
}
