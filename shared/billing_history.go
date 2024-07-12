package shared

import "encoding/json"

type BillingHistory struct {
	User               string       `json:"user"`
	Date               Date         `json:"date"`
	Event              BillingEvent `json:"event"`
	OutstandingBalance float64      `json:"outstanding_balance"`
}

type BillingEvent interface {
	GetType() BillingEventType
}

func (b *BillingHistory) UnmarshalJSON(data []byte) (err error) {
	var typ struct {
		Event struct {
			Type BillingEventType `json:"type"`
		} `json:"event"`
	}
	if err := json.Unmarshal(data, &typ); err != nil {
		return err
	}
	switch typ.Event.Type {
	case EventTypeInvoiceGenerated:
		b.Event = new(InvoiceGenerated)
	case EventTypeFeeReductionAwarded:
		b.Event = new(FeeReductionAwarded)
	case EventTypeFeeReductionApplied:
		b.Event = new(FeeReductionApplied)
	case EventTypeInvoiceAdjustmentApproved:
		b.Event = new(InvoiceAdjustmentApproved)
	case EventTypeInvoiceAdjustmentPending:
		b.Event = new(InvoiceAdjustmentPending)
	case EventTypePaymentProcessed:
		b.Event = new(PaymentProcessed)
	default:
		// ignore
	}
	type tmp BillingHistory // avoids infinite recursion
	err = json.Unmarshal(data, (*tmp)(b))
	return err
}

type InvoiceEvent struct {
	ID        int    `json:"id"`
	Reference string `json:"reference"`
}

type BaseBillingEvent struct {
	Type BillingEventType `json:"type"`
}

func (d BaseBillingEvent) GetType() BillingEventType {
	return d.Type
}

type InvoiceGenerated struct {
	ClientId         string       `json:"client_id"`
	InvoiceReference InvoiceEvent `json:"invoice_reference"`
	InvoiceType      string       `json:"invoice_type"`
	InvoiceName      string       `json:"invoice_name"`
	Amount           string       `json:"amount"`
	BaseBillingEvent
}

type FeeReductionAwarded struct {
	ReductionType string `json:"reduction_type"`
	StartDate     Date   `json:"start_date"`
	EndDate       Date   `json:"end_date"`
	DateReceived  Date   `json:"date_received"`
	Notes         string `json:"notes"`
	BaseBillingEvent
}

type FeeReductionApplied struct {
	ClientId         string `json:"client_id"`
	ReductionType    string `json:"reduction_type"`
	PaymentBreakdown `json:"payment_breakdown"`
	BaseBillingEvent
}

type InvoiceAdjustmentApproved struct {
	AdjustmentType   string `json:"adjustment_type"`
	ClientId         string `json:"client_id"`
	Notes            string `json:"notes"`
	PaymentBreakdown `json:"payment_breakdown"`
	BaseBillingEvent
}

type InvoiceAdjustmentPending struct {
	AdjustmentType   string `json:"adjustment_type"`
	ClientId         string `json:"client_id"`
	Notes            string `json:"notes"`
	PaymentBreakdown `json:"payment_breakdown"`
	BaseBillingEvent
}

type PaymentProcessed struct {
	PaymentType string             `json:"payment_type"`
	Total       int                `json:"total"`
	Breakdown   []PaymentBreakdown `json:"breakdown"`
	BaseBillingEvent
}

type PaymentBreakdown struct {
	InvoiceReference InvoiceEvent `json:"invoice_reference"`
	Amount           int          `json:"amount"`
}

type BillingEventType int

const (
	EventTypeUnknown BillingEventType = iota
	EventTypeInvoiceGenerated
	EventTypeFeeReductionAwarded
	EventTypeFeeReductionApplied
	EventTypeInvoiceAdjustmentApproved
	EventTypePaymentProcessed
	EventTypeInvoiceAdjustmentPending
)

var eventTypeMap = map[string]BillingEventType{
	"UNKNOWN":                    EventTypeUnknown,
	"INVOICE_GENERATED":          EventTypeInvoiceGenerated,
	"FEE_REDUCTION_AWARDED":      EventTypeFeeReductionAwarded,
	"FEE_REDUCTION_APPLIED":      EventTypeFeeReductionApplied,
	"INVOICE_ADJUSTMENT_APPLIED": EventTypeInvoiceAdjustmentApproved,
	"INVOICE_ADJUSTMENT_PENDING": EventTypeInvoiceAdjustmentPending,
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
	case EventTypeInvoiceAdjustmentPending:
		return "INVOICE_ADJUSTMENT_PENDING"
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
