package server

import (
	"net/http"
	"strconv"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type Refunds []Refund

type Refund struct {
	ID            string
	DateRaised    shared.Date
	DateFulfilled *shared.Date
	Amount        int
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
	Refunds       Refunds
	ShowAddRefund bool
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

	data := &RefundsTab{Refunds: h.transform(refunds), ShowAddRefund: h.shouldShowAddRefund(refunds), ClientId: strconv.Itoa(clientID), AppVars: v}
	data.selectTab("refunds")

	return h.execute(w, r, data)
}

func (h *RefundsHandler) shouldShowAddRefund(refunds shared.Refunds) bool {
	if refunds.CreditBalance == 0 {
		return false
	}
	for _, r := range refunds.Refunds {
		switch r.Status {
		case shared.RefundStatusPending,
			shared.RefundStatusApproved,
			shared.RefundStatusProcessing:
			return false
		default:
			continue
		}
	}
	return true
}

func (h *RefundsHandler) transform(in shared.Refunds) Refunds {
	var out Refunds
	for _, r := range in.Refunds {
		refund := Refund{
			ID:         strconv.Itoa(r.ID),
			DateRaised: r.RaisedDate,
			Amount:     r.Amount,
			Notes:      r.Notes,
			CreatedBy:  r.CreatedBy,
			Status:     r.Status.String(),
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
