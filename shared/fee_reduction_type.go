package shared

import "encoding/json"

type FeeReductionType int

const (
	FeeReductionTypeUnknown FeeReductionType = iota
	FeeReductionTypeExemption
	FeeReductionTypeHardship
	FeeReductionTypeRemission
)

var feeReductionTypeMap = map[string]FeeReductionType{
	"EXEMPTION": FeeReductionTypeExemption,
	"HARDSHIP":  FeeReductionTypeHardship,
	"REMISSION": FeeReductionTypeRemission,
	// also match to equivalent transaction type keys
	"CREDIT EXEMPTION": FeeReductionTypeExemption,
	"CREDIT HARDSHIP":  FeeReductionTypeHardship,
	"CREDIT REMISSION": FeeReductionTypeRemission,
}

func (f FeeReductionType) String() string {
	switch f {
	case FeeReductionTypeExemption:
		return "Exemption"
	case FeeReductionTypeHardship:
		return "Hardship"
	case FeeReductionTypeRemission:
		return "Remission"
	default:
		return ""
	}
}

func (f FeeReductionType) Key() string {
	switch f {
	case FeeReductionTypeExemption:
		return "EXEMPTION"
	case FeeReductionTypeHardship:
		return "HARDSHIP"
	case FeeReductionTypeRemission:
		return "REMISSION"
	default:
		return ""
	}
}

func ParseFeeReductionType(s string) FeeReductionType {
	value, ok := feeReductionTypeMap[s]
	if !ok {
		return FeeReductionType(0)
	}
	return value
}

func (f FeeReductionType) Valid() bool {
	return f != FeeReductionTypeUnknown
}

func (f FeeReductionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Key())
}

func (f *FeeReductionType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*f = ParseFeeReductionType(s)
	return nil
}
