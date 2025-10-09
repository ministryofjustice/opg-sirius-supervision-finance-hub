package shared

import "encoding/json"

type Refunds struct {
	Refunds       []Refund `json:"refunds"`
	CreditBalance int      `json:"creditBalance"`
}

type Refund struct {
	ID            int                   `json:"id"`
	RaisedDate    Date                  `json:"raisedDate"`
	FulfilledDate Nillable[Date]        `json:"fulfilledDate"`
	Amount        int                   `json:"amount"`
	Status        RefundStatus          `json:"status"`
	Notes         string                `json:"notes"`
	BankDetails   Nillable[BankDetails] `json:"bankDetails"`
	CreatedBy     int                   `json:"createdBy"`
}

type BankDetails struct {
	Name     string `json:"name" validate:"required"`
	Account  string `json:"account" validate:"required,numeric,len=8"`
	SortCode string `json:"sortCode" validate:"required,len=8"`
}

type AddRefund struct {
	AccountName   string `json:"name" validate:"required"`
	AccountNumber string `json:"account" validate:"required,numeric,len=8"`
	SortCode      string `json:"sortCode" validate:"required,len=8"`
	RefundNotes   string `json:"notes" validate:"required,thousand-character-limit"`
}

type UpdateRefundStatus struct {
	Status RefundStatus `json:"status" validate:"valid-enum,oneof=2 3 5"` // APPROVED, REJECTED, CANCELLED
}

type RefundStatus int

const (
	RefundStatusUnknown RefundStatus = iota
	RefundStatusPending
	RefundStatusApproved
	RefundStatusRejected
	RefundStatusProcessing
	RefundStatusCancelled
	RefundStatusFulfilled
)

var refundStatusMap = map[string]RefundStatus{
	"PENDING":    RefundStatusPending,
	"APPROVED":   RefundStatusApproved,
	"REJECTED":   RefundStatusRejected,
	"PROCESSING": RefundStatusProcessing,
	"CANCELLED":  RefundStatusCancelled,
	"FULFILLED":  RefundStatusFulfilled,
}

func (i RefundStatus) String() string {
	switch i {
	case RefundStatusPending:
		return "Pending"
	case RefundStatusApproved:
		return "Approved"
	case RefundStatusRejected:
		return "Rejected"
	case RefundStatusProcessing:
		return "Processing"
	case RefundStatusCancelled:
		return "Cancelled"
	case RefundStatusFulfilled:
		return "Fulfilled"
	default:
		return ""
	}
}

func (i RefundStatus) Key() string {
	switch i {
	case RefundStatusPending:
		return "PENDING"
	case RefundStatusApproved:
		return "APPROVED"
	case RefundStatusRejected:
		return "REJECTED"
	case RefundStatusProcessing:
		return "PROCESSING"
	case RefundStatusCancelled:
		return "CANCELLED"
	case RefundStatusFulfilled:
		return "FULFILLED"
	default:
		return ""
	}
}

func (i RefundStatus) Valid() bool {
	return i != RefundStatusUnknown
}

func ParseRefundStatus(s string) RefundStatus {
	value, ok := refundStatusMap[s]
	if !ok {
		return RefundStatus(0)
	}
	return value
}

func (i RefundStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Key())
}

func (i *RefundStatus) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*i = ParseRefundStatus(s)
	return nil
}
