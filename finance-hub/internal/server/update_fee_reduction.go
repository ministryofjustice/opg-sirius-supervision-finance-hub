package server

import (
	"net/http"
)

type UpdateFeeReductions struct {
	ClientId string
	AppVars
}

type UpdateFeeReductionHandler struct {
	router
}

func (h *UpdateFeeReductionHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	data := UpdateFeeReductions{r.PathValue("clientId"), v}

	return h.execute(w, r, data)
}
