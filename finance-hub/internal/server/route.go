package server

import (
	"github.com/opg-sirius-finance-hub/finance-hub/internal/util"
	"github.com/opg-sirius-finance-hub/shared"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"net/http"
)

type FinanceClient struct {
	FirstName          string
	Surname            string
	CourtRef           string
	OutstandingBalance string
	CreditBalance      string
	PaymentMethod      string
}

type HeaderData struct {
	MyDetails     shared.Assignee
	FinanceClient FinanceClient
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

		var person shared.Person
		var accountInfo shared.AccountInformation

		group.Go(func() error {
			myDetails, err := r.client.GetCurrentUserDetails(ctx.With(groupCtx))
			if err != nil {
				return err
			}
			data.MyDetails = myDetails
			return nil
		})
		group.Go(func() error {
			p, err := r.client.GetPersonDetails(ctx.With(groupCtx), ctx.ClientId)
			if err != nil {
				return err
			}
			person = p
			return nil
		})
		group.Go(func() error {
			ai, err := r.client.GetAccountInformation(ctx.With(groupCtx), ctx.ClientId)
			if err != nil {
				return err
			}
			accountInfo = ai
			return nil
		})

		if err := group.Wait(); err != nil {
			return err
		}

		data.FinanceClient = r.transformFinanceClient(person, accountInfo)
		data.SuccessMessage = r.getSuccess(req)

		return r.tmpl.Execute(w, data)
	}
}

func (r route) getSuccess(req *http.Request) string {
	switch req.URL.Query().Get("success") {
	case "CREDIT_WRITE_OFF":
		return "The write off is now waiting for approval"
	case "remission":
		return "The remission has been successfully added"
	case "exemption":
		return "The exemption has been successfully added"
	case "hardship":
		return "The hardship has been successfully added"
	case "feeReductionCancelled":
		return "The fee reduction has been successfully cancelled"
	case "credit":
		return "You have approved the " + req.URL.Query().Get("success")
	case "write off":
		return "You have approved the " + req.URL.Query().Get("success")
	}
	return ""
}

func (r route) isHxRequest(req *http.Request) bool {
	return req.Header.Get("HX-Request") == "true"
}

func (r route) transformFinanceClient(person shared.Person, accountInfo shared.AccountInformation) FinanceClient {
	return FinanceClient{
		FirstName:          person.FirstName,
		Surname:            person.Surname,
		CourtRef:           person.CourtRef,
		OutstandingBalance: util.IntToDecimalString(accountInfo.OutstandingBalance),
		CreditBalance:      util.IntToDecimalString(accountInfo.CreditBalance),
		PaymentMethod:      cases.Title(language.English).String(accountInfo.PaymentMethod),
	}
}
