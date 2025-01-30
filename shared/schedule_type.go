package shared

import "encoding/json"

type ScheduleType int

const (
	ScheduleTypeUnknown ScheduleType = iota
	ScheduleTypeMOTOCardPayments
	ScheduleTypeOnlineCardPayments
	ScheduleTypeOPGBACSTransfer
	ScheduleTypeSupervisionBACSTransfer
	ScheduleTypeDirectDebitPayments
	ScheduleTypeAdFeeInvoices
	ScheduleTypeS2FeeInvoices
	ScheduleTypeS3FeeInvoices
	ScheduleTypeB2FeeInvoices
	ScheduleTypeB3FeeInvoices
	ScheduleTypeSFFeeInvoicesGeneral
	ScheduleTypeSFFeeInvoicesMinimal
	ScheduleTypeSEFeeInvoicesGeneral
	ScheduleTypeSEFeeInvoicesMinimal
	ScheduleTypeSOFeeInvoicesGeneral
	ScheduleTypeSOFeeInvoicesMinimal
	ScheduleTypeADFeeReductions
	ScheduleTypeGeneralManualCredits
	ScheduleTypeMinimalManualCredits
	ScheduleTypeGeneralManualDebits
	ScheduleTypeMinimalManualDebits
	ScheduleTypeADWriteOffs
	ScheduleTypeGeneralWriteOffs
	ScheduleTypeMinimalWriteOffs
	ScheduleTypeADWriteOffReversals
	ScheduleTypeGeneralWriteOffReversals
	ScheduleTypeMinimalWriteOffReversals
)

var scheduleTypeMap = map[string]ScheduleType{
	"MOTOCardPayments":         ScheduleTypeMOTOCardPayments,
	"OnlineCardPayments":       ScheduleTypeOnlineCardPayments,
	"OPGBACSTransfer":          ScheduleTypeOPGBACSTransfer,
	"SupervisionBACSTransfer":  ScheduleTypeSupervisionBACSTransfer,
	"DirectDebitPayment":       ScheduleTypeDirectDebitPayments,
	"AdFeeInvoices":            ScheduleTypeAdFeeInvoices,
	"S2FeeInvoices":            ScheduleTypeS2FeeInvoices,
	"S3FeeInvoices":            ScheduleTypeS3FeeInvoices,
	"B2FeeInvoices":            ScheduleTypeB2FeeInvoices,
	"B3FeeInvoices":            ScheduleTypeB3FeeInvoices,
	"SFFeeInvoicesGeneral":     ScheduleTypeSFFeeInvoicesGeneral,
	"SFFeeInvoicesMinimal":     ScheduleTypeSFFeeInvoicesMinimal,
	"SEFeeInvoicesGeneral":     ScheduleTypeSEFeeInvoicesGeneral,
	"SEFeeInvoicesMinimal":     ScheduleTypeSEFeeInvoicesMinimal,
	"SOFeeInvoicesGeneral":     ScheduleTypeSOFeeInvoicesGeneral,
	"SOFeeInvoicesMinimal":     ScheduleTypeSOFeeInvoicesMinimal,
	"ADFeeReductions":          ScheduleTypeADFeeReductions,
	"GeneralManualCredits":     ScheduleTypeGeneralManualCredits,
	"MinimalManualCredits":     ScheduleTypeMinimalManualCredits,
	"GeneralManualDebits":      ScheduleTypeGeneralManualDebits,
	"MinimalManualDebits":      ScheduleTypeMinimalManualDebits,
	"ADWrite-offs":             ScheduleTypeADWriteOffs,
	"GeneralWrite-offs":        ScheduleTypeGeneralWriteOffs,
	"MinimalWriteOffs":         ScheduleTypeMinimalWriteOffs,
	"ADWriteOffReversals":      ScheduleTypeADWriteOffReversals,
	"GeneralWriteOffReversals": ScheduleTypeGeneralWriteOffReversals,
	"MinimalWriteOffReversals": ScheduleTypeMinimalWriteOffReversals,
}

func (s ScheduleType) String() string {
	return s.Key()
}

func (s ScheduleType) Translation() string {
	switch s {
	case ScheduleTypeMOTOCardPayments:
		return "MOTO Card Payments"
	case ScheduleTypeOnlineCardPayments:
		return "Online Card Payments"
	case ScheduleTypeOPGBACSTransfer:
		return "OPG BACS Transfer"
	case ScheduleTypeSupervisionBACSTransfer:
		return "Supervision BACS transfer"
	case ScheduleTypeDirectDebitPayments:
		return "Direct Debit Payment"
	case ScheduleTypeAdFeeInvoices:
		return "Ad Fee Invoices"
	case ScheduleTypeS2FeeInvoices:
		return "S2 Fee Invoices"
	case ScheduleTypeS3FeeInvoices:
		return "S3 Fee Invoices"
	case ScheduleTypeB2FeeInvoices:
		return "B2 Fee Invoices"
	case ScheduleTypeB3FeeInvoices:
		return "B3 Fee Invoices"
	case ScheduleTypeSFFeeInvoicesGeneral:
		return "SF Fee Invoices (General)"
	case ScheduleTypeSFFeeInvoicesMinimal:
		return "SF Fee Invoices (Minimal)"
	case ScheduleTypeSEFeeInvoicesGeneral:
		return "SE Fee Invoices (General)"
	case ScheduleTypeSEFeeInvoicesMinimal:
		return "SE Fee Invoices (Minimal)"
	case ScheduleTypeSOFeeInvoicesGeneral:
		return "SO Fee Invoices (General)"
	case ScheduleTypeSOFeeInvoicesMinimal:
		return "SO Fee Invoices (Minimal)"
	case ScheduleTypeADFeeReductions:
		return "AD Fee Reductions"
	case ScheduleTypeGeneralManualCredits:
		return "General Manual Credits"
	case ScheduleTypeMinimalManualCredits:
		return "Minimal Manual Credits"
	case ScheduleTypeGeneralManualDebits:
		return "General Manual Debits"
	case ScheduleTypeMinimalManualDebits:
		return "Minimal Manual Debits"
	case ScheduleTypeADWriteOffs:
		return "AD Write-offs"
	case ScheduleTypeGeneralWriteOffs:
		return "General Write-offs"
	case ScheduleTypeMinimalWriteOffs:
		return "Minimal Write-offs"
	case ScheduleTypeADWriteOffReversals:
		return "AD Write-off Reversals"
	case ScheduleTypeGeneralWriteOffReversals:
		return "General Write-off Reversals"
	case ScheduleTypeMinimalWriteOffReversals:
		return "Minimal Write-off Reversals"
	default:
		return ""
	}
}

func (s ScheduleType) Key() string {
	switch s {
	case ScheduleTypeMOTOCardPayments:
		return "MOTOCardPayments"
	case ScheduleTypeOnlineCardPayments:
		return "OnlineCardPayments"
	case ScheduleTypeOPGBACSTransfer:
		return "OPGBACSTransfer"
	case ScheduleTypeSupervisionBACSTransfer:
		return "SupervisionBACSTransfer"
	case ScheduleTypeDirectDebitPayments:
		return "DirectDebitPayment"
	case ScheduleTypeAdFeeInvoices:
		return "AdFeeInvoices"
	case ScheduleTypeS2FeeInvoices:
		return "S2FeeInvoices"
	case ScheduleTypeS3FeeInvoices:
		return "S3FeeInvoices"
	case ScheduleTypeB2FeeInvoices:
		return "B2FeeInvoices"
	case ScheduleTypeB3FeeInvoices:
		return "B3FeeInvoices"
	case ScheduleTypeSFFeeInvoicesGeneral:
		return "SFFeeInvoicesGeneral"
	case ScheduleTypeSFFeeInvoicesMinimal:
		return "SFFeeInvoicesMinimal"
	case ScheduleTypeSEFeeInvoicesGeneral:
		return "SEFeeInvoicesGeneral"
	case ScheduleTypeSEFeeInvoicesMinimal:
		return "SEFeeInvoicesMinimal"
	case ScheduleTypeSOFeeInvoicesGeneral:
		return "SOFeeInvoicesGeneral"
	case ScheduleTypeSOFeeInvoicesMinimal:
		return "SOFeeInvoicesMinimal"
	case ScheduleTypeADFeeReductions:
		return "ADFeeReductions"
	case ScheduleTypeGeneralManualCredits:
		return "GeneralManualCredits"
	case ScheduleTypeMinimalManualCredits:
		return "MinimalManualCredits"
	case ScheduleTypeGeneralManualDebits:
		return "GeneralManualDebits"
	case ScheduleTypeMinimalManualDebits:
		return "MinimalManualDebits"
	case ScheduleTypeADWriteOffs:
		return "ADWrite-offs"
	case ScheduleTypeGeneralWriteOffs:
		return "GeneralWrite-offs"
	case ScheduleTypeMinimalWriteOffs:
		return "MinimalWriteOffs"
	case ScheduleTypeADWriteOffReversals:
		return "ADWriteOffReversals"
	case ScheduleTypeGeneralWriteOffReversals:
		return "GeneralWriteOffReversals"
	case ScheduleTypeMinimalWriteOffReversals:
		return "MinimalWriteOffReversals"
	default:
		return ""
	}
}

func ParseScheduleType(s string) ScheduleType {
	value, ok := scheduleTypeMap[s]
	if !ok {
		return ScheduleType(0)
	}
	return value
}

func (s ScheduleType) Valid() bool {
	return s != ScheduleTypeUnknown
}

func (s ScheduleType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Key())
}

func (s *ScheduleType) UnmarshalJSON(data []byte) (err error) {
	var st string
	if err := json.Unmarshal(data, &st); err != nil {
		return err
	}
	*s = ParseScheduleType(st)
	return nil
}
