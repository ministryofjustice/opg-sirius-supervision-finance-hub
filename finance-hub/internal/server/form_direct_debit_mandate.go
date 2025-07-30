package server

import (
	"net/http"
)

type DirectDebitMandateForm struct {
	ClientId string
	AppVars
}

type DirectDebitMandateHandler struct {
	router
}

func (h *DirectDebitMandateHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	data := DirectDebitMandateForm{r.PathValue("clientId"), v}

	return h.execute(w, r, data)
}
