package server

import (
	"fmt"
	"net/http"
)

type UpdateDirectDebit struct {
	ClientId string
	AppVars
}

type UpdateDirectDebitHandler struct {
	router
}

func (h *UpdateDirectDebitHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	fmt.Println("form direct debit handler")

	data := UpdateDirectDebit{r.PathValue("clientId"), v}

	return h.execute(w, r, data)
}
