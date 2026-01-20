package shared

import (
	"encoding/json"
)

type PaymentMethod int

const (
	PaymentMethodUnknown PaymentMethod = iota
	PaymentMethodDemanded
	PaymentMethodDirectDebit
)

var paymentMethodMap = map[string]PaymentMethod{
	"DEMANDED":     PaymentMethodDemanded,
	"DIRECT DEBIT": PaymentMethodDirectDebit,
}

func (i PaymentMethod) String() string {
	switch i {
	case PaymentMethodDemanded:
		return "Demanded"
	case PaymentMethodDirectDebit:
		return "Direct Debit"
	default:
		return ""
	}
}

func (i PaymentMethod) Key() string {
	switch i {
	case PaymentMethodDemanded:
		return "DEMANDED"
	case PaymentMethodDirectDebit:
		return "DIRECT DEBIT"
	default:
		return ""
	}
}

func (i PaymentMethod) Valid() bool {
	return i != PaymentMethodUnknown
}

func ParsePaymentMethod(s string) PaymentMethod {
	value, ok := paymentMethodMap[s]
	if !ok {
		return PaymentMethod(0)
	}
	return value
}

func (i PaymentMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Key())
}

func (i *PaymentMethod) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*i = ParsePaymentMethod(s)
	return nil
}
