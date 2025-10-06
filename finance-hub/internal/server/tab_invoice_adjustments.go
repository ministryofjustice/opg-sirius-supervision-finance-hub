package server

import (
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"math"
	"net/http"
	"strconv"
)

type InvoiceAdjustments []PendingInvoiceAdjustment

type PendingInvoiceAdjustment struct {
	Id               string
	Invoice          string
	DateRaised       shared.Date
	AdjustmentType   string
	AdjustmentAmount string
	Notes            string
	Status           string
	CreatedBy        int
}

type InvoiceAdjustmentsTab struct {
	InvoiceAdjustments InvoiceAdjustments
	ClientId           string
	AppVars
}

type InvoiceAdjustmentsHandler struct {
	router
}

func (h *InvoiceAdjustmentsHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	clientID := getClientID(r)

	ia, err := h.Client().GetInvoiceAdjustments(ctx, clientID)
	if err != nil {
		return err
	}

	data := &InvoiceAdjustmentsTab{InvoiceAdjustments: h.transform(ia), ClientId: strconv.Itoa(clientID), AppVars: v}
	data.selectTab("invoice-adjustments")
	fmt.Printf("in invoice-adjustments render")
	return h.execute(w, r, data)
}

func (h *InvoiceAdjustmentsHandler) transform(in shared.InvoiceAdjustments) InvoiceAdjustments {
	var out InvoiceAdjustments
	for _, adjustment := range in {
		out = append(out, PendingInvoiceAdjustment{
			Id:               strconv.Itoa(adjustment.Id),
			Invoice:          adjustment.InvoiceRef,
			DateRaised:       adjustment.RaisedDate,
			AdjustmentType:   h.transformType(adjustment.AdjustmentType),
			AdjustmentAmount: shared.IntToDecimalString(int(math.Abs(float64(adjustment.Amount)))),
			Notes:            adjustment.Notes,
			Status:           h.transformStatus(adjustment.Status),
			CreatedBy:        adjustment.CreatedBy,
		})
	}

	return out
}

func (h *InvoiceAdjustmentsHandler) transformType(in shared.AdjustmentType) string {
	switch in {
	case shared.AdjustmentTypeCreditMemo:
		return "Credit"
	case shared.AdjustmentTypeDebitMemo:
		return "Debit"
	case shared.AdjustmentTypeWriteOff:
		return "Write off"
	case shared.AdjustmentTypeWriteOffReversal:
		return "Write off reversal"
	case shared.AdjustmentTypeFeeReductionReversal:
		return "Fee reduction reversal"
	default:
		return ""
	}
}

func (h *InvoiceAdjustmentsHandler) transformStatus(in string) string {
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
