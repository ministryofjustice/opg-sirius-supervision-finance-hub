package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"net/http"
)

type WithHeaderData struct {
	Data any
	HeaderData
}

type route struct {
	client  *ApiClient
	tmpl    Template
	partial string
	Data    any
}

func (r route) execute(w http.ResponseWriter, req *http.Request) error {
	if r.isHxRequest(req) {
		return r.tmpl.ExecuteTemplate(w, r.partial, r.Data)
	} else {
		data := WithHeaderData{
			Data: r.Data,
			HeaderData: HeaderData{
				MyDetails: model.Assignee{},
				Client: ClientVars{
					FirstName:   "Ian",
					Surname:     "Moneybags",
					Outstanding: "1000000",
				},
			},
		}
		return r.tmpl.Execute(w, data)
	}
}

func (r route) isHxRequest(req *http.Request) bool {
	return req.Header.Get("HX-Request") == "true"
}
