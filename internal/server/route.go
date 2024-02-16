package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"golang.org/x/sync/errgroup"
	"net/http"
)

type HeaderData struct {
	MyDetails model.Assignee
	Client    model.Person
}

type PageData struct {
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
		ctx := getContext(req)
		group, groupCtx := errgroup.WithContext(ctx.Context)

		data := PageData{
			Data: r.Data,
		}

		group.Go(func() error {
			myDetails, err := (*r.client).GetCurrentUserDetails(ctx.With(groupCtx))
			if err != nil {
				return err
			}
			data.MyDetails = myDetails
			return nil
		})
		group.Go(func() error {
			//re := regexp.MustCompile("[0-9]+")
			personId := 1
			person, err := (*r.client).GetPersonDetails(ctx.With(groupCtx), personId)
			if err != nil {
				return err
			}
			data.Client = person
			return nil
		})

		if err := group.Wait(); err != nil {
			return err
		}

		return r.tmpl.Execute(w, data)
	}
}

func (r route) isHxRequest(req *http.Request) bool {
	return req.Header.Get("HX-Request") == "true"
}
