package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"golang.org/x/sync/errgroup"
	"net/http"
)

type HeaderData struct {
	MyDetails model.Assignee
	Person    model.Person
}

type PageData struct {
	Data           any
	SuccessMessage string
	HeaderData
}

type route struct {
	client  ApiClient
	tmpl    Template
	partial string
}

func (r route) Client() ApiClient {
	return r.client
}

// execute is an abstraction of the Template execute functions in order to conditionally render either a full template or
// a block, in response to a header added by HTMX. If the header is not present, the function will also fetch all
// additional data needed by the page for a full page load.
func (r route) execute(w http.ResponseWriter, req *http.Request, data any) error {
	if r.isHxRequest(req) {
		return r.tmpl.ExecuteTemplate(w, r.partial, data)
	} else {
		ctx := getContext(req)
		group, groupCtx := errgroup.WithContext(ctx.Context)

		data := PageData{
			Data: data,
		}

		group.Go(func() error {
			myDetails, err := r.client.GetCurrentUserDetails(ctx.With(groupCtx))
			if err != nil {
				return err
			}
			data.MyDetails = myDetails
			return nil
		})
		group.Go(func() error {
			person, err := r.client.GetPersonDetails(ctx.With(groupCtx), ctx.ClientId)
			if err != nil {
				return err
			}
			data.Person = person
			return nil
		})

		if err := group.Wait(); err != nil {
			return err
		}

		data.SuccessMessage = r.getSuccess(req)

		return r.tmpl.Execute(w, data)
	}
}

func (r route) getSuccess(req *http.Request) string {
	switch req.URL.Query().Get("success") {
	case "writeOff":
		return "The write off is now waiting for approval"
	}
	return ""
}

func (r route) isHxRequest(req *http.Request) bool {
	return req.Header.Get("HX-Request") == "true"
}
