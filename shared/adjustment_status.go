package shared

import (
	"encoding/json"
)

type AdjustmentStatus int

const (
	AdjustmentStatusUnknown AdjustmentStatus = iota
	AdjustmentStatusPending
	AdjustmentStatusApproved
	AdjustmentStatusRejected
)

var adjustmentStatusMap = map[string]AdjustmentStatus{
	"PENDING":  AdjustmentStatusPending,
	"APPROVED": AdjustmentStatusApproved,
	"REJECTED": AdjustmentStatusRejected,
}

func (i AdjustmentStatus) String() string {
	switch i {
	case AdjustmentStatusPending:
		return "Pending"
	case AdjustmentStatusApproved:
		return "Approved"
	case AdjustmentStatusRejected:
		return "Rejected"
	default:
		return ""
	}
}

func (i AdjustmentStatus) Key() string {
	switch i {
	case AdjustmentStatusPending:
		return "PENDING"
	case AdjustmentStatusApproved:
		return "APPROVED"
	case AdjustmentStatusRejected:
		return "REJECTED"
	default:
		return ""
	}
}

func (i AdjustmentStatus) Valid() bool {
	return i != AdjustmentStatusUnknown
}

func ParseAdjustmentStatus(s string) AdjustmentStatus {
	value, ok := adjustmentStatusMap[s]
	if !ok {
		return AdjustmentStatus(0)
	}
	return value
}

func (i AdjustmentStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Key())
}

func (i *AdjustmentStatus) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*i = ParseAdjustmentStatus(s)
	return nil
}
