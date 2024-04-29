package server

import (
	"github.com/opg-sirius-finance-hub/finance-hub/internal/util"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

type PendingInvoiceAdjustments []PendingInvoiceAdjustment

type PendingInvoiceAdjustment struct {
	Id               int
	Invoice          string
	DateRaised       shared.Date
	AdjustmentType   string
	AdjustmentAmount string
	Notes            string
	Status           string
}

type PendingInvoiceAdjustmentsTab struct {
	PendingInvoiceAdjustments PendingInvoiceAdjustments
	AppVars
}

type PendingInvoiceAdjustmentsHandler struct {
	router
}

func (h *PendingInvoiceAdjustmentsHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)

	ia, err := h.Client().GetInvoiceAdjustments(ctx, ctx.ClientId)
	if err != nil {
		return err
	}

	data := &PendingInvoiceAdjustmentsTab{PendingInvoiceAdjustments: h.transform(ia), AppVars: v}
	data.selectTab("pending-invoice-adjustments")

	return h.execute(w, r, data)
}

func (h *PendingInvoiceAdjustmentsHandler) transform(in shared.InvoiceAdjustments) PendingInvoiceAdjustments {
	var out PendingInvoiceAdjustments
	for _, adjustment := range in {
		out = append(out, PendingInvoiceAdjustment{
			Id:               adjustment.Id,
			Invoice:          adjustment.InvoiceRef,
			DateRaised:       adjustment.RaisedDate,
			AdjustmentType:   h.transformType(adjustment.AdjustmentType),
			AdjustmentAmount: util.IntToDecimalString(adjustment.Amount),
			Notes:            adjustment.Notes,
			Status:           h.transformStatus(adjustment.Status),
		})
	}

	return out
}

func (h *PendingInvoiceAdjustmentsHandler) transformType(in string) string {
	if in == "CREDIT MEMO" {
		return "Credit"
	} else if in == "CREDIT WRITE OFF" {
		return "Write off"
	} else {
		return ""
	}
}

func (h *PendingInvoiceAdjustmentsHandler) transformStatus(in string) string {
	if in == "PENDING" {
		return "Pending"
	} else if in == "REJECTED" {
		return "Rejected"
	} else if in == "APPROVED" {
		return "Approved"
	} else {
		return ""
	}
}
