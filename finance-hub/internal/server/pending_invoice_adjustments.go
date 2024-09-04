package server

import (
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
)

type PendingInvoiceAdjustments []PendingInvoiceAdjustment

type PendingInvoiceAdjustment struct {
	Id               string
	Invoice          string
	DateRaised       shared.Date
	AdjustmentType   string
	AdjustmentAmount string
	Notes            string
	Status           string
}

type PendingInvoiceAdjustmentsTab struct {
	PendingInvoiceAdjustments PendingInvoiceAdjustments
	ClientId                  string
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

	data := &PendingInvoiceAdjustmentsTab{PendingInvoiceAdjustments: h.transform(ia), ClientId: strconv.Itoa(ctx.ClientId), AppVars: v}
	data.selectTab("pending-invoice-adjustments")

	return h.execute(w, r, data)
}

func (h *PendingInvoiceAdjustmentsHandler) transform(in shared.InvoiceAdjustments) PendingInvoiceAdjustments {
	var out PendingInvoiceAdjustments
	for _, adjustment := range in {
		out = append(out, PendingInvoiceAdjustment{
			Id:               strconv.Itoa(adjustment.Id),
			Invoice:          adjustment.InvoiceRef,
			DateRaised:       adjustment.RaisedDate,
			AdjustmentType:   h.transformType(adjustment.AdjustmentType),
			AdjustmentAmount: shared.IntToDecimalString(adjustment.Amount),
			Notes:            adjustment.Notes,
			Status:           h.transformStatus(adjustment.Status),
		})
	}

	return out
}

func (h *PendingInvoiceAdjustmentsHandler) transformType(in shared.AdjustmentType) string {
	switch in {
	case shared.AdjustmentTypeCreditMemo:
		return "Credit"
	case shared.AdjustmentTypeDebitMemo:
		return "Debit"
	case shared.AdjustmentTypeWriteOff:
		return "Write off"
	case shared.AdjustmentTypeWriteOffReversal:
		return "Write off reversal"
	default:
		return ""
	}
}

func (h *PendingInvoiceAdjustmentsHandler) transformStatus(in string) string {
	switch in {
	case "PENDING":
		return "Pending"
	case "REJECTED":
		return "Rejected"
	case "APPROVED":
		return "Approved"
	default:
		return ""
	}
}
