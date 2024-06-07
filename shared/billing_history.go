package shared

import "encoding/json"

type BillingHistory struct {
	User               string       `json:"user"`
	Date               Date         `json:"date"`
	Event              BillingEvent `json:"event"`
	OutstandingBalance int          `json:"outstanding_balance"`
}

type BillingEvent struct {
	Type BillingEventType `json:"type"`
	Data interface{}      `json:"data"`
}

type InvoiceReference struct {
	ID        int    `json:"id"`
	Reference string `json:"reference"`
}

type InvoiceGenerated struct {
	InvoiceReference InvoiceReference `json:"invoice_reference"`
	InvoiceType      string           `json:"invoice_type"`
	Amount           int              `json:"amount"`
}

type FeeReductionAwarded struct {
	ReductionType string `json:"reduction_type"`
	StartDate     Date   `json:"start_date"`
	EndDate       Date   `json:"end_date"`
	DateReceived  Date   `json:"date_received"`
	Notes         string `json:"notes"`
}

type FeeReductionApplied struct {
	ReductionType string `json:"reduction_type"`
	PaymentBreakdown
}

type InvoiceAdjustmentApproved struct {
	AdjustmentType string `json:"adjustment_type"`
	PaymentBreakdown
}

type PaymentProcessed struct {
	PaymentType string             `json:"payment_type"`
	Total       int                `json:"total"`
	Breakdown   []PaymentBreakdown `json:"breakdown"`
}

type PaymentBreakdown struct {
	InvoiceReference InvoiceReference `json:"invoice_reference"`
	Amount           int              `json:"amount"`
}

type BillingEventType int

const (
	EventTypeUnknown BillingEventType = iota
	EventTypeInvoiceGenerated
	EventTypeFeeReductionAwarded
	EventTypeFeeReductionApplied
	EventTypeInvoiceAdjustmentApproved
	EventTypePaymentProcessed
)

var eventTypeMap = map[string]BillingEventType{
	"UNKNOWN":                    EventTypeUnknown,
	"INVOICE_GENERATED":          EventTypeInvoiceGenerated,
	"FEE_REDUCTION_AWARDED":      EventTypeFeeReductionAwarded,
	"FEE_REDUCTION_APPLIED":      EventTypeFeeReductionApplied,
	"INVOICE_ADJUSTMENT_APPLIED": EventTypeInvoiceAdjustmentApproved,
	"PAYMENT_PROCESSED":          EventTypePaymentProcessed,
}

func (b BillingEventType) String() string {
	switch b {
	case EventTypeInvoiceGenerated:
		return "INVOICE_GENERATED"
	case EventTypeFeeReductionAwarded:
		return "FEE_REDUCTION_AWARDED"
	case EventTypeFeeReductionApplied:
		return "FEE_REDUCTION_APPLIED"
	case EventTypeInvoiceAdjustmentApproved:
		return "INVOICE_ADJUSTMENT_APPLIED"
	case EventTypePaymentProcessed:
		return "PAYMENT_PROCESSED"
	default:
		return "UNKNOWN"
	}
}

func (b BillingEventType) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.String())
}

func (b *BillingEventType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	e, ok := eventTypeMap[s]
	if !ok {
		*b = EventTypeUnknown
	} else {
		*b = e
	}
	return nil
}
