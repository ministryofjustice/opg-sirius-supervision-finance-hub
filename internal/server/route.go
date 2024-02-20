package server

import (
	"errors"
	"github.com/opg-sirius-finance-hub/internal/model"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
)

type HeaderData struct {
	MyDetails model.Assignee
	Person    model.Person
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

// execute is an abstraction of the Template execute functions in order to conditionally render either a full template or
// a block, in response to a header added by HTMX. If the header is not present, the function will also fetch all
// additional data needed by the page for a full page load.
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
			personId, err := strconv.Atoi(req.PathValue("id"))
			if err != nil {
				return errors.New("client id in string cannot be parsed to an integer")
			}
			person, err := (*r.client).GetPersonDetails(ctx.With(groupCtx), personId)
			if err != nil {
				return err
			}
			data.Person = person
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
