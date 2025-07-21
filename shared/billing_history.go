package shared

import (
	"encoding/json"
)

type BillingHistory struct {
	User               int          `json:"user"`
	Date               Date         `json:"date"`
	Event              BillingEvent `json:"event"`
	OutstandingBalance int          `json:"outstanding_balance"`
	CreditBalance      int          `json:"credit_balance"`
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
	case EventTypeFeeReductionCancelled:
		b.Event = new(FeeReductionCancelled)
	case EventTypeFeeReductionApplied:
		b.Event = new(FeeReductionApplied)
	case EventTypeInvoiceAdjustmentApplied:
		b.Event = new(InvoiceAdjustmentApplied)
	case EventTypeInvoiceAdjustmentPending:
		b.Event = new(InvoiceAdjustmentPending)
	case EventTypeInvoiceAdjustmentRejected:
		b.Event = new(InvoiceAdjustmentRejected)
	case EventTypePaymentProcessed:
		b.Event = new(PaymentProcessed)
	case EventTypeReappliedCredit:
		b.Event = new(ReappliedCredit)
	default:
		b.Event = new(UnknownEvent)
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

type UnknownEvent struct {
	BaseBillingEvent
}

type InvoiceGenerated struct {
	ClientId         int          `json:"client_id"`
	InvoiceReference InvoiceEvent `json:"invoice_reference"`
	InvoiceType      InvoiceType  `json:"invoice_type"`
	Amount           int          `json:"amount"`
	BaseBillingEvent
}

type FeeReductionAwarded struct {
	ReductionType FeeReductionType `json:"reduction_type"`
	StartDate     Date             `json:"start_date"`
	EndDate       Date             `json:"end_date"`
	DateReceived  Date             `json:"date_received"`
	Notes         string           `json:"notes"`
	BaseBillingEvent
}

type FeeReductionCancelled struct {
	ReductionType      FeeReductionType `json:"reduction_type"`
	CancellationReason string           `json:"cancellation_reason"`
	BaseBillingEvent
}

type InvoiceAdjustmentPending struct {
	AdjustmentType   AdjustmentType `json:"adjustment_type"`
	ClientId         int            `json:"client_id"`
	Notes            string         `json:"notes"`
	PaymentBreakdown `json:"payment_breakdown"`
	BaseBillingEvent
}

type InvoiceAdjustmentRejected struct {
	AdjustmentType   AdjustmentType `json:"adjustment_type"`
	ClientId         int            `json:"client_id"`
	Notes            string         `json:"notes"`
	PaymentBreakdown `json:"payment_breakdown"`
	BaseBillingEvent
}

type TransactionEvent struct {
	ClientId        int                `json:"client_id"`
	TransactionType TransactionType    `json:"transaction_type"`
	Amount          int                `json:"amount"` // the amount that triggered the transaction, excluding unapplies
	Breakdown       []PaymentBreakdown `json:"breakdown"`
	BaseBillingEvent
}

type PaymentBreakdown struct {
	InvoiceReference InvoiceEvent `json:"invoice_reference"`
	Amount           int          `json:"amount"`
	Status           string       `json:"status"`
}

type InvoiceAdjustmentApplied struct {
	TransactionEvent
}

type FeeReductionApplied struct {
	TransactionEvent
}

type PaymentProcessed struct {
	TransactionEvent
}

type ReappliedCredit struct {
	TransactionEvent
}

type BillingEventType int

const (
	EventTypeUnknown BillingEventType = iota
	EventTypeInvoiceGenerated
	EventTypeFeeReductionAwarded
	EventTypeFeeReductionCancelled
	EventTypeFeeReductionApplied
	EventTypeInvoiceAdjustmentApplied
	EventTypePaymentProcessed
	EventTypeInvoiceAdjustmentPending
	EventTypeInvoiceAdjustmentRejected
	EventTypeReappliedCredit
)

var eventTypeMap = map[string]BillingEventType{
	"UNKNOWN":                     EventTypeUnknown,
	"INVOICE_GENERATED":           EventTypeInvoiceGenerated,
	"FEE_REDUCTION_AWARDED":       EventTypeFeeReductionAwarded,
	"FEE_REDUCTION_CANCELLED":     EventTypeFeeReductionCancelled,
	"FEE_REDUCTION_APPLIED":       EventTypeFeeReductionApplied,
	"INVOICE_ADJUSTMENT_APPLIED":  EventTypeInvoiceAdjustmentApplied,
	"INVOICE_ADJUSTMENT_PENDING":  EventTypeInvoiceAdjustmentPending,
	"INVOICE_ADJUSTMENT_REJECTED": EventTypeInvoiceAdjustmentRejected,
	"PAYMENT_PROCESSED":           EventTypePaymentProcessed,
	"REAPPLIED_CREDIT":            EventTypeReappliedCredit,
}

func (b BillingEventType) String() string {
	switch b {
	case EventTypeInvoiceGenerated:
		return "INVOICE_GENERATED"
	case EventTypeFeeReductionAwarded:
		return "FEE_REDUCTION_AWARDED"
	case EventTypeFeeReductionApplied:
		return "FEE_REDUCTION_APPLIED"
	case EventTypeFeeReductionCancelled:
		return "FEE_REDUCTION_CANCELLED"
	case EventTypeInvoiceAdjustmentApplied:
		return "INVOICE_ADJUSTMENT_APPLIED"
	case EventTypeInvoiceAdjustmentPending:
		return "INVOICE_ADJUSTMENT_PENDING"
	case EventTypeInvoiceAdjustmentRejected:
		return "INVOICE_ADJUSTMENT_REJECTED"
	case EventTypePaymentProcessed:
		return "PAYMENT_PROCESSED"
	case EventTypeReappliedCredit:
		return "REAPPLIED_CREDIT"
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
