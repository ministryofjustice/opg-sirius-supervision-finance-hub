package server

import (
	"net/http"
)

type AddFeeReductionForm struct {
	ClientId string
	AppVars
}

type AddFeeReductionHandler struct {
	router
}

func (h *AddFeeReductionHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	data := AddFeeReductionForm{r.PathValue("clientId"), v}

	return h.execute(w, r, data)
}
