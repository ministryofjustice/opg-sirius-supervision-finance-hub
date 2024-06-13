package shared

import (
	"encoding/json"
)

var InvoiceTypes = []InvoiceType{
	InvoiceTypeAD,
	InvoiceTypeS2,
	InvoiceTypeS3,
	InvoiceTypeB2,
	InvoiceTypeB3,
	InvoiceTypeSF,
	InvoiceTypeSE,
	InvoiceTypeSO,
}

type InvoiceType int

const (
	InvoiceTypeUnknown InvoiceType = iota
	InvoiceTypeAD
	InvoiceTypeS2
	InvoiceTypeS3
	InvoiceTypeB2
	InvoiceTypeB3
	InvoiceTypeSF
	InvoiceTypeSE
	InvoiceTypeSO
)

var invoiceTypeMap = map[string]InvoiceType{
	"AD": InvoiceTypeAD,
	"S2": InvoiceTypeS2,
	"S3": InvoiceTypeS3,
	"B2": InvoiceTypeB2,
	"B3": InvoiceTypeB3,
	"SF": InvoiceTypeSF,
	"SE": InvoiceTypeSE,
	"SO": InvoiceTypeSO,
}

func (i InvoiceType) Translation() string {
	switch i {
	case InvoiceTypeAD:
		return "AD - Assessment of deputy fee invoice"
	case InvoiceTypeS2:
		return "S2- General annual fee invoice (Non-direct debit)"
	case InvoiceTypeS3:
		return "S3 - Minimal annual fee invoice (Non-direct debit)"
	case InvoiceTypeB2:
		return "B2 - General annual fee invoice (Direct debit)"
	case InvoiceTypeB3:
		return "B3 - Minimal annual fee invoice (Direct debit)"
	case InvoiceTypeSF:
		return "SF - Client deceased final fee invoice"
	case InvoiceTypeSE:
		return "SE - Full order expired final fee invoice"
	case InvoiceTypeSO:
		return "SO - Client regained capacity final fee invoice}"
	default:
		return ""
	}
}

func (i InvoiceType) Key() string {
	switch i {
	case InvoiceTypeAD:
		return "AD"
	case InvoiceTypeS2:
		return "S2"
	case InvoiceTypeS3:
		return "S3"
	case InvoiceTypeB2:
		return "B2"
	case InvoiceTypeB3:
		return "B3"
	case InvoiceTypeSF:
		return "SF"
	case InvoiceTypeSE:
		return "SE"
	case InvoiceTypeSO:
		return "SO"
	default:
		return ""
	}
}

func ParseInvoiceType(s string) InvoiceType {
	value, ok := invoiceTypeMap[s]
	if !ok {
		return InvoiceType(0)
	}
	return value
}

func (i InvoiceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Key())
}

func (i *InvoiceType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*i = ParseInvoiceType(s)
	return nil
}
