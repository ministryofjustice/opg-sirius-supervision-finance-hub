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
	DateFulfilled *shared.Date
	Amount        string
	BankDetails   *BankDetails
	Notes         string
	CreatedBy     int
	Status        string
}

type BankDetails struct {
	Name     string
	Account  string
	SortCode string
}

type RefundsTab struct {
	Refunds       []Refund
	CreditBalance int
	ClientId      string
	AppVars
}

type RefundsHandler struct {
	router
}

func (h *RefundsHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	clientID := getClientID(r)

	refunds, err := h.Client().GetRefunds(ctx, clientID)
	if err != nil {
		return err
	}

	data := &RefundsTab{Refunds: h.transform(refunds), CreditBalance: refunds.CreditBalance, ClientId: strconv.Itoa(clientID), AppVars: v}
	data.selectTab("refunds")

	return h.execute(w, r, data)
}

func (h *RefundsHandler) transform(in shared.Refunds) Refunds {
	var out Refunds
	for _, r := range in.Refunds {
		refund := Refund{
			ID:         strconv.Itoa(r.ID),
			DateRaised: r.RaisedDate,
			Amount:     shared.IntToDecimalString(r.Amount),
			Notes:      r.Notes,
			CreatedBy:  r.CreatedBy,
			Status:     h.transformStatus(r.Status),
		}
		if r.FulfilledDate.Valid {
			refund.DateFulfilled = &r.FulfilledDate.Value
		}
		if r.BankDetails.Valid {
			bd := r.BankDetails.Value
			refund.BankDetails = &BankDetails{
				Name:     bd.Name,
				Account:  bd.Account,
				SortCode: bd.SortCode,
			}
		}
		out = append(out, refund)
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
	case "FULFILLED":
		return "Fulfilled"
	default:
		return ""
	}
}
