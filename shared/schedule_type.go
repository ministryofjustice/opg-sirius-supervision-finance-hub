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
	ScheduleTypeChequePayments
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
	ScheduleTypeGAFeeInvoices
	ScheduleTypeGSFeeInvoices
	ScheduleTypeGTFeeInvoices
	ScheduleTypeADFeeReductions
	ScheduleTypeGeneralFeeReductions
	ScheduleTypeMinimalFeeReductions
	ScheduleTypeGAFeeReductions
	ScheduleTypeGSFeeReductions
	ScheduleTypeGTFeeReductions
	ScheduleTypeADManualCredits
	ScheduleTypeGeneralManualCredits
	ScheduleTypeMinimalManualCredits
	ScheduleTypeGAManualCredits
	ScheduleTypeGSManualCredits
	ScheduleTypeGTManualCredits
	ScheduleTypeADManualDebits
	ScheduleTypeGeneralManualDebits
	ScheduleTypeMinimalManualDebits
	ScheduleTypeGAManualDebits
	ScheduleTypeGSManualDebits
	ScheduleTypeGTManualDebits
	ScheduleTypeADWriteOffs
	ScheduleTypeGeneralWriteOffs
	ScheduleTypeMinimalWriteOffs
	ScheduleTypeGAWriteOffs
	ScheduleTypeGSWriteOffs
	ScheduleTypeGTWriteOffs
	ScheduleTypeADWriteOffReversals
	ScheduleTypeGeneralWriteOffReversals
	ScheduleTypeMinimalWriteOffReversals
	ScheduleTypeGAWriteOffReversals
	ScheduleTypeGSWriteOffReversals
	ScheduleTypeGTWriteOffReversals
	ScheduleTypeUnappliedPayments
	ScheduleTypeReappliedPayments
	ScheduleTypeRefunds
	ScheduleTypeADFeeReductionReversals
	ScheduleTypeGeneralFeeReductionReversals
	ScheduleTypeMinimalFeeReductionReversals
	ScheduleTypeGAFeeReductionReversals
	ScheduleTypeGSFeeReductionReversals
	ScheduleTypeGTFeeReductionReversals
)

var scheduleTypeMap = map[string]ScheduleType{
	"MOTOCardPayments":             ScheduleTypeMOTOCardPayments,
	"OnlineCardPayments":           ScheduleTypeOnlineCardPayments,
	"OPGBACSTransfer":              ScheduleTypeOPGBACSTransfer,
	"SupervisionBACSTransfer":      ScheduleTypeSupervisionBACSTransfer,
	"DirectDebitPayment":           ScheduleTypeDirectDebitPayments,
	"ChequePayments":               ScheduleTypeChequePayments,
	"AdFeeInvoices":                ScheduleTypeAdFeeInvoices,
	"S2FeeInvoices":                ScheduleTypeS2FeeInvoices,
	"S3FeeInvoices":                ScheduleTypeS3FeeInvoices,
	"B2FeeInvoices":                ScheduleTypeB2FeeInvoices,
	"B3FeeInvoices":                ScheduleTypeB3FeeInvoices,
	"SFFeeInvoicesGeneral":         ScheduleTypeSFFeeInvoicesGeneral,
	"SFFeeInvoicesMinimal":         ScheduleTypeSFFeeInvoicesMinimal,
	"SEFeeInvoicesGeneral":         ScheduleTypeSEFeeInvoicesGeneral,
	"SEFeeInvoicesMinimal":         ScheduleTypeSEFeeInvoicesMinimal,
	"SOFeeInvoicesGeneral":         ScheduleTypeSOFeeInvoicesGeneral,
	"SOFeeInvoicesMinimal":         ScheduleTypeSOFeeInvoicesMinimal,
	"GAFeeInvoices":                ScheduleTypeGAFeeInvoices,
	"GSFeeInvoices":                ScheduleTypeGSFeeInvoices,
	"GTFeeInvoices":                ScheduleTypeGTFeeInvoices,
	"ADFeeReductions":              ScheduleTypeADFeeReductions,
	"GeneralFeeReductions":         ScheduleTypeGeneralFeeReductions,
	"MinimalFeeReductions":         ScheduleTypeMinimalFeeReductions,
	"GAFeeReductions":              ScheduleTypeGAFeeReductions,
	"GSFeeReductions":              ScheduleTypeGSFeeReductions,
	"GTFeeReductions":              ScheduleTypeGTFeeReductions,
	"ADManualCredits":              ScheduleTypeADManualCredits,
	"GeneralManualCredits":         ScheduleTypeGeneralManualCredits,
	"MinimalManualCredits":         ScheduleTypeMinimalManualCredits,
	"GAManualCredits":              ScheduleTypeGAManualCredits,
	"GSManualCredits":              ScheduleTypeGSManualCredits,
	"GTManualCredits":              ScheduleTypeGTManualCredits,
	"ADManualDebits":               ScheduleTypeADManualDebits,
	"GeneralManualDebits":          ScheduleTypeGeneralManualDebits,
	"MinimalManualDebits":          ScheduleTypeMinimalManualDebits,
	"GAManualDebits":               ScheduleTypeGAManualDebits,
	"GSManualDebits":               ScheduleTypeGSManualDebits,
	"GTManualDebits":               ScheduleTypeGTManualDebits,
	"ADWriteOffs":                  ScheduleTypeADWriteOffs,
	"GeneralWriteOffs":             ScheduleTypeGeneralWriteOffs,
	"MinimalWriteOffs":             ScheduleTypeMinimalWriteOffs,
	"GAWriteOffs":                  ScheduleTypeGAWriteOffs,
	"GSWriteOffs":                  ScheduleTypeGSWriteOffs,
	"GTWriteOffs":                  ScheduleTypeGTWriteOffs,
	"ADWriteOffReversals":          ScheduleTypeADWriteOffReversals,
	"GeneralWriteOffReversals":     ScheduleTypeGeneralWriteOffReversals,
	"MinimalWriteOffReversals":     ScheduleTypeMinimalWriteOffReversals,
	"GAWriteOffReversals":          ScheduleTypeGAWriteOffReversals,
	"GSWriteOffReversals":          ScheduleTypeGSWriteOffReversals,
	"GTWriteOffReversals":          ScheduleTypeGTWriteOffReversals,
	"UnappliedPayments":            ScheduleTypeUnappliedPayments,
	"ReappliedPayments":            ScheduleTypeReappliedPayments,
	"Refunds":                      ScheduleTypeRefunds,
	"ADFeeReductionReversals":      ScheduleTypeADFeeReductionReversals,
	"GeneralFeeReductionReversals": ScheduleTypeGeneralFeeReductionReversals,
	"MinimalFeeReductionReversals": ScheduleTypeMinimalFeeReductionReversals,
	"GAFeeReductionReversals":      ScheduleTypeGAFeeReductionReversals,
	"GSFeeReductionReversals":      ScheduleTypeGSFeeReductionReversals,
	"GTFeeReductionReversals":      ScheduleTypeGTFeeReductionReversals,
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
	case ScheduleTypeChequePayments:
		return "Cheque Payments"
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
	case ScheduleTypeGAFeeInvoices:
		return "GA Fee Invoices"
	case ScheduleTypeGSFeeInvoices:
		return "GS Fee Invoices"
	case ScheduleTypeGTFeeInvoices:
		return "GT Fee Invoices"
	case ScheduleTypeADFeeReductions:
		return "AD Fee Reductions"
	case ScheduleTypeGeneralFeeReductions:
		return "General Fee Reductions"
	case ScheduleTypeMinimalFeeReductions:
		return "Minimal Fee Reductions"
	case ScheduleTypeGAFeeReductions:
		return "GA Fee Reductions"
	case ScheduleTypeGSFeeReductions:
		return "GS Fee Reductions"
	case ScheduleTypeGTFeeReductions:
		return "GT Fee Reductions"
	case ScheduleTypeADManualCredits:
		return "AD Manual Credits"
	case ScheduleTypeGeneralManualCredits:
		return "General Manual Credits"
	case ScheduleTypeMinimalManualCredits:
		return "Minimal Manual Credits"
	case ScheduleTypeGAManualCredits:
		return "GA Manual Credits"
	case ScheduleTypeGSManualCredits:
		return "GS Manual Credits"
	case ScheduleTypeGTManualCredits:
		return "GT Manual Credits"
	case ScheduleTypeADManualDebits:
		return "AD Manual Debits"
	case ScheduleTypeGeneralManualDebits:
		return "General Manual Debits"
	case ScheduleTypeMinimalManualDebits:
		return "Minimal Manual Debits"
	case ScheduleTypeGAManualDebits:
		return "GA Manual Debits"
	case ScheduleTypeGSManualDebits:
		return "GS Manual Debits"
	case ScheduleTypeGTManualDebits:
		return "GT Manual Debits"
	case ScheduleTypeADWriteOffs:
		return "AD Write-offs"
	case ScheduleTypeGeneralWriteOffs:
		return "General Write-offs"
	case ScheduleTypeMinimalWriteOffs:
		return "Minimal Write-offs"
	case ScheduleTypeGAWriteOffs:
		return "GA Write-offs"
	case ScheduleTypeGSWriteOffs:
		return "GS Write-offs"
	case ScheduleTypeGTWriteOffs:
		return "GT Write-offs"
	case ScheduleTypeADWriteOffReversals:
		return "AD Write-off Reversals"
	case ScheduleTypeGeneralWriteOffReversals:
		return "General Write-off Reversals"
	case ScheduleTypeMinimalWriteOffReversals:
		return "Minimal Write-off Reversals"
	case ScheduleTypeGAWriteOffReversals:
		return "GA Write-off Reversals"
	case ScheduleTypeGSWriteOffReversals:
		return "GS Write-off Reversals"
	case ScheduleTypeGTWriteOffReversals:
		return "GT Write-off Reversals"
	case ScheduleTypeUnappliedPayments:
		return "Unapplied payments"
	case ScheduleTypeReappliedPayments:
		return "Reapplied payments"
	case ScheduleTypeRefunds:
		return "Refunds"
	case ScheduleTypeADFeeReductionReversals:
		return "AD Fee Reduction Reversals"
	case ScheduleTypeGeneralFeeReductionReversals:
		return "General Fee Reduction Reversals"
	case ScheduleTypeMinimalFeeReductionReversals:
		return "Minimal Fee Reduction Reversals"
	case ScheduleTypeGAFeeReductionReversals:
		return "GA Fee Reduction Reversals"
	case ScheduleTypeGSFeeReductionReversals:
		return "GS Fee Reduction Reversals"
	case ScheduleTypeGTFeeReductionReversals:
		return "GT Fee Reduction Reversals"
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
	case ScheduleTypeChequePayments:
		return "ChequePayments"
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
	case ScheduleTypeGAFeeInvoices:
		return "GAFeeInvoices"
	case ScheduleTypeGSFeeInvoices:
		return "GSFeeInvoices"
	case ScheduleTypeGTFeeInvoices:
		return "GTFeeInvoices"
	case ScheduleTypeADFeeReductions:
		return "ADFeeReductions"
	case ScheduleTypeGeneralFeeReductions:
		return "GeneralFeeReductions"
	case ScheduleTypeMinimalFeeReductions:
		return "MinimalFeeReductions"
	case ScheduleTypeGAFeeReductions:
		return "GAFeeReductions"
	case ScheduleTypeGSFeeReductions:
		return "GSFeeReductions"
	case ScheduleTypeGTFeeReductions:
		return "GTFeeReductions"
	case ScheduleTypeADManualCredits:
		return "ADManualCredits"
	case ScheduleTypeGeneralManualCredits:
		return "GeneralManualCredits"
	case ScheduleTypeMinimalManualCredits:
		return "MinimalManualCredits"
	case ScheduleTypeGAManualCredits:
		return "GAManualCredits"
	case ScheduleTypeGSManualCredits:
		return "GSManualCredits"
	case ScheduleTypeGTManualCredits:
		return "GTManualCredits"
	case ScheduleTypeADManualDebits:
		return "ADManualDebits"
	case ScheduleTypeGeneralManualDebits:
		return "GeneralManualDebits"
	case ScheduleTypeMinimalManualDebits:
		return "MinimalManualDebits"
	case ScheduleTypeGAManualDebits:
		return "GAManualDebits"
	case ScheduleTypeGSManualDebits:
		return "GSManualDebits"
	case ScheduleTypeGTManualDebits:
		return "GTManualDebits"
	case ScheduleTypeADWriteOffs:
		return "ADWriteOffs"
	case ScheduleTypeGeneralWriteOffs:
		return "GeneralWriteOffs"
	case ScheduleTypeMinimalWriteOffs:
		return "MinimalWriteOffs"
	case ScheduleTypeGAWriteOffs:
		return "GAWriteOffs"
	case ScheduleTypeGSWriteOffs:
		return "GSWriteOffs"
	case ScheduleTypeGTWriteOffs:
		return "GTWriteOffs"
	case ScheduleTypeADWriteOffReversals:
		return "ADWriteOffReversals"
	case ScheduleTypeGeneralWriteOffReversals:
		return "GeneralWriteOffReversals"
	case ScheduleTypeMinimalWriteOffReversals:
		return "MinimalWriteOffReversals"
	case ScheduleTypeGAWriteOffReversals:
		return "GAWriteOffReversals"
	case ScheduleTypeGSWriteOffReversals:
		return "GSWriteOffReversals"
	case ScheduleTypeGTWriteOffReversals:
		return "GTWriteOffReversals"
	case ScheduleTypeUnappliedPayments:
		return "UnappliedPayments"
	case ScheduleTypeReappliedPayments:
		return "ReappliedPayments"
	case ScheduleTypeRefunds:
		return "Refunds"
	case ScheduleTypeADFeeReductionReversals:
		return "ADFeeReductionReversals"
	case ScheduleTypeGeneralFeeReductionReversals:
		return "GeneralFeeReductionReversals"
	case ScheduleTypeMinimalFeeReductionReversals:
		return "MinimalFeeReductionReversals"
	case ScheduleTypeGAFeeReductionReversals:
		return "GAFeeReductionReversals"
	case ScheduleTypeGSFeeReductionReversals:
		return "GSFeeReductionReversals"
	case ScheduleTypeGTFeeReductionReversals:
		return "GTFeeReductionReversals"
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
