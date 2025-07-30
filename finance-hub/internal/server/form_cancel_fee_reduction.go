package server

import (
	"net/http"
)

type CancelFeeReductionForm struct {
	ClientId string
	Id       string
	AppVars
}

type CancelFeeReductionHandler struct {
	router
}

func (h *CancelFeeReductionHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	data := CancelFeeReductionForm{r.PathValue("clientId"), r.PathValue("feeReductionId"), v}

	return h.execute(w, r, data)
}
