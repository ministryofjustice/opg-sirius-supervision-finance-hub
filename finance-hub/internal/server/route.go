package server

import (
	"net/http"
	"strconv"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type FinanceClient struct {
	FirstName          string
	Surname            string
	CourtRef           string
	OutstandingBalance int
	CreditBalance      int
	PaymentMethod      string
	ClientId           string
}

type HeaderData struct {
	User          *shared.User
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
	if IsHxRequest(req) {
		return r.tmpl.ExecuteTemplate(w, r.partial, data)
	} else {
		ctx := req.Context().(auth.Context)
		group, groupCtx := errgroup.WithContext(ctx)
		ctx = ctx.WithContext(groupCtx)

		data := PageData{
			Data: data,
		}

		data.User = ctx.User

		clientID := getClientID(req)
		var person shared.Person
		var accountInfo shared.AccountInformation

		group.Go(func() error {
			p, err := r.client.GetPersonDetails(ctx, clientID)
			if err != nil {
				return err
			}
			person = p
			return nil
		})
		group.Go(func() error {
			ai, err := r.client.GetAccountInformation(ctx, clientID)
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
	case "invoice-adjustment[CREDIT WRITE OFF]":
		return "Write-off successfully created"
	case "invoice-adjustment[CREDIT MEMO]":
		return "Manual credit successfully created"
	case "invoice-adjustment[DEBIT MEMO]":
		return "Manual debit successfully created"
	case "invoice-adjustment[WRITE OFF REVERSAL]":
		return "Write-off reversal successfully created"
	case "invoice-adjustment[FEE REDUCTION REVERSAL]":
		return "Fee reduction reversal successfully created"
	case "fee-reduction[REMISSION]":
		return "The remission has been successfully added"
	case "fee-reduction[EXEMPTION]":
		return "The exemption has been successfully added"
	case "fee-reduction[HARDSHIP]":
		return "The hardship has been successfully added"
	case "fee-reduction[CANCELLED]":
		return "The fee reduction has been successfully cancelled"
	case "approved-invoice-adjustment[CREDIT]":
		return "You have approved the credit"
	case "approved-invoice-adjustment[WRITE OFF]":
		return "You have approved the write off"
	case "approved-invoice-adjustment[WRITE OFF REVERSAL]":
		return "You have approved the write off reversal"
	case "approved-invoice-adjustment[DEBIT]":
		return "You have approved the debit"
	case "approved-invoice-adjustment[FEE REDUCTION REVERSAL]":
		return "You have approved the fee reduction reversal"
	case "rejected-invoice-adjustment[CREDIT]":
		return "You have rejected the credit"
	case "rejected-invoice-adjustment[WRITE OFF]":
		return "You have rejected the write off"
	case "rejected-invoice-adjustment[WRITE OFF REVERSAL]":
		return "You have rejected the write off reversal"
	case "rejected-invoice-adjustment[DEBIT]":
		return "You have rejected the debit"
	case "rejected-invoice-adjustment[FEE REDUCTION REVERSAL]":
		return "You have rejected the fee reduction reversal"
	case "invoice-type[AD]":
		return "The AD invoice has been successfully created"
	case "invoice-type[S2]":
		return "The S2 invoice has been successfully created"
	case "invoice-type[S3]":
		return "The S3 invoice has been successfully created"
	case "invoice-type[B2]":
		return "The B2 invoice has been successfully created"
	case "invoice-type[B3]":
		return "The B3 invoice has been successfully created"
	case "invoice-type[SF]":
		return "The SF invoice has been successfully created"
	case "invoice-type[SE]":
		return "The SE invoice has been successfully created"
	case "invoice-type[SO]":
		return "The SO invoice has been successfully created"
	case "invoice-type[GA]":
		return "The GA invoice has been successfully created"
	case "invoice-type[GS]":
		return "The GS invoice has been successfully created"
	case "invoice-type[GT]":
		return "The GT invoice has been successfully created"
	case "payment-method":
		return "Payment method has been successfully changed"
	case "refund-added":
		return "The refund has been successfully added"
	case "refunds[APPROVED]":
		return "You have approved the refund"
	case "refunds[REJECTED]":
		return "You have rejected the refund"
	case "refunds[CANCELLED]":
		return "You have cancelled the refund"
	case "direct-debit":
		return "The Direct Debit has been set up"
	case "cancel-direct-debit":
		return "The Direct Debit has been cancelled"
	}
	return ""
}

func IsHxRequest(req *http.Request) bool {
	return req.Header.Get("HX-Request") == "true"
}

func (r route) transformFinanceClient(person shared.Person, accountInfo shared.AccountInformation) FinanceClient {
	return FinanceClient{
		ClientId:           strconv.Itoa(person.ID),
		FirstName:          person.FirstName,
		Surname:            person.Surname,
		CourtRef:           person.CourtRef,
		OutstandingBalance: accountInfo.OutstandingBalance,
		CreditBalance:      accountInfo.CreditBalance,
		PaymentMethod:      cases.Title(language.English).String(accountInfo.PaymentMethod),
	}
}

func getClientID(req *http.Request) int {
	clientId, _ := strconv.Atoi(req.PathValue("clientId"))
	return clientId
}
