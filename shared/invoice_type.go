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
	InvoiceTypeGA,
	InvoiceTypeGS,
	InvoiceTypeGT,
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
	InvoiceTypeGA
	InvoiceTypeGS
	InvoiceTypeGT
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
	"GA": InvoiceTypeGA,
	"GS": InvoiceTypeGS,
	"GT": InvoiceTypeGT,
}

func (i InvoiceType) String() string {
	return i.Key()
}

func (i InvoiceType) Translation() string {
	switch i {
	case InvoiceTypeAD:
		return "AD - Assessment of deputy fee invoice"
	case InvoiceTypeS2:
		return "S2 - General annual fee invoice (Non-Direct Debit)"
	case InvoiceTypeS3:
		return "S3 - Minimal annual fee invoice (Non-Direct Debit)"
	case InvoiceTypeB2:
		return "B2 - General annual fee invoice (Direct Debit)"
	case InvoiceTypeB3:
		return "B3 - Minimal annual fee invoice (Direct Debit)"
	case InvoiceTypeSF:
		return "SF - Client deceased final fee invoice"
	case InvoiceTypeSE:
		return "SE - Full order expired final fee invoice"
	case InvoiceTypeSO:
		return "SO - Client regained capacity final fee invoice"
	case InvoiceTypeGA:
		return "GA - Assessment of Guardian"
	case InvoiceTypeGS:
		return "GS - Guardianship supervision invoice"
	case InvoiceTypeGT:
		return "GT - Guardianship termination invoice"
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
	case InvoiceTypeGA:
		return "GA"
	case InvoiceTypeGS:
		return "GS"
	case InvoiceTypeGT:
		return "GT"
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

func (i InvoiceType) Valid() bool {
	return i != InvoiceTypeUnknown
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

func (i InvoiceType) RequiresDateValidation() bool {
	switch i {
	case InvoiceTypeAD, InvoiceTypeSF, InvoiceTypeSE, InvoiceTypeSO, InvoiceTypeGA, InvoiceTypeGT, InvoiceTypeGS:
		return true
	default:
		return false
	}
}

func (i InvoiceType) RequiresSameFinancialYearValidation() bool {
	switch i {
	case InvoiceTypeSF, InvoiceTypeSE, InvoiceTypeSO:
		return true
	default:
		return false
	}
}
