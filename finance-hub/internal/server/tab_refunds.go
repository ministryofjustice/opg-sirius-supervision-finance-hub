package server

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
	"strconv"
)

type Refunds []Refund

type Refund struct {
	ID            string
	DateRaised    shared.Date
	DateFulfilled shared.Date
	Amount        string
	BankDetails   BankDetails
	Notes         string
	Status        string
}

type BankDetails struct {
	Name     string
	Account  string
	SortCode string
}

type RefundsTab struct {
	Refunds  []Refund
	ClientId string
	AppVars
}

type RefundsHandler struct {
	router
}

func (h *RefundsHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	clientID := getClientID(r)

	ia, err := h.Client().GetRefunds(ctx, clientID)
	if err != nil {
		return err
	}

	data := &RefundsTab{Refunds: h.transform(ia), ClientId: strconv.Itoa(clientID), AppVars: v}
	data.selectTab("refunds")

	return h.execute(w, r, data)
}

func (h *RefundsHandler) transform(in shared.Refunds) Refunds {
	var out Refunds
	for _, refund := range in {
		out = append(out, Refund{
			ID:            strconv.Itoa(refund.ID),
			DateRaised:    refund.RaisedDate,
			DateFulfilled: refund.FulfilledDate,
			Amount:        shared.IntToDecimalString(refund.Amount),
			BankDetails: BankDetails{
				Name:     refund.BankDetails.Name,
				Account:  refund.BankDetails.Account,
				SortCode: refund.BankDetails.SortCode,
			},
			Notes:  refund.Notes,
			Status: h.transformStatus(refund.Status),
		})
	}

	return out
}

func (h *RefundsHandler) transformStatus(in string) string {
	switch in {
	case "PENDING":
		return "Pending"
	case "REJECTED":
		return "Rejected"
	case "APPROVED":
		return "Approved"
	case "PROCESSING":
		return "Processing"
	case "CANCELLED":
		return "Cancelled"
	default:
		return ""
	}
}
