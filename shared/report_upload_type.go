package shared

import (
	"encoding/json"
)

var reportUploadPaymentTypes = []ReportUploadType{
	ReportTypeUploadPaymentsMOTOCard,
	ReportTypeUploadPaymentsOnlineCard,
	ReportTypeUploadPaymentsOPGBACS,
	ReportTypeUploadPaymentsSupervisionBACS,
	ReportTypeUploadPaymentsSupervisionCheque,
	ReportTypeUploadDirectDebitsCollections,
}

var reportUploadReversalTypes = []ReportUploadType{
	ReportTypeUploadMisappliedPayments,
}

var reportUploadNoHeaderTypes = []ReportUploadType{
	ReportTypeUploadDirectDebitsCollections,
}

type ReportUploadType int

const (
	ReportTypeUploadUnknown ReportUploadType = iota
	ReportTypeUploadPaymentsMOTOCard
	ReportTypeUploadPaymentsOnlineCard
	ReportTypeUploadPaymentsOPGBACS
	ReportTypeUploadPaymentsSupervisionBACS
	ReportTypeUploadPaymentsSupervisionCheque
	ReportTypeUploadDebtChase
	ReportTypeUploadDeputySchedule
	ReportTypeUploadDirectDebitsCollections
	ReportTypeUploadMisappliedPayments
)

var reportTypeUploadMap = map[string]ReportUploadType{
	"PAYMENTS_MOTO_CARD":          ReportTypeUploadPaymentsMOTOCard,
	"PAYMENTS_ONLINE_CARD":        ReportTypeUploadPaymentsOnlineCard,
	"PAYMENTS_OPG_BACS":           ReportTypeUploadPaymentsOPGBACS,
	"PAYMENTS_SUPERVISION_BACS":   ReportTypeUploadPaymentsSupervisionBACS,
	"PAYMENTS_SUPERVISION_CHEQUE": ReportTypeUploadPaymentsSupervisionCheque,
	"DEBT_CHASE":                  ReportTypeUploadDebtChase,
	"DEPUTY_SCHEDULE":             ReportTypeUploadDeputySchedule,
	"DIRECT_DEBITS_COLLECTIONS":   ReportTypeUploadDirectDebitsCollections,
	"MISAPPLIED_PAYMENTS":         ReportTypeUploadMisappliedPayments,
}

func (u ReportUploadType) String() string {
	return u.Key()
}

func (u ReportUploadType) Translation() string {
	switch u {
	case ReportTypeUploadPaymentsMOTOCard:
		return "Payments - MOTO card"
	case ReportTypeUploadPaymentsOnlineCard:
		return "Payments - Online card"
	case ReportTypeUploadPaymentsOPGBACS:
		return "Payments - OPG BACS"
	case ReportTypeUploadPaymentsSupervisionBACS:
		return "Payments - Supervision BACS"
	case ReportTypeUploadPaymentsSupervisionCheque:
		return "Payments - Supervision Cheque"
	case ReportTypeUploadDebtChase:
		return "Debt chase"
	case ReportTypeUploadDeputySchedule:
		return "Deputy schedule"
	case ReportTypeUploadDirectDebitsCollections:
		return "Direct debits collections"
	case ReportTypeUploadMisappliedPayments:
		return "Payment Reversals - Misapplied payments"
	default:
		return ""
	}
}

func (u ReportUploadType) Key() string {
	switch u {
	case ReportTypeUploadPaymentsMOTOCard:
		return "PAYMENTS_MOTO_CARD"
	case ReportTypeUploadPaymentsOnlineCard:
		return "PAYMENTS_ONLINE_CARD"
	case ReportTypeUploadPaymentsOPGBACS:
		return "PAYMENTS_OPG_BACS"
	case ReportTypeUploadPaymentsSupervisionBACS:
		return "PAYMENTS_SUPERVISION_BACS"
	case ReportTypeUploadPaymentsSupervisionCheque:
		return "PAYMENTS_SUPERVISION_CHEQUE"
	case ReportTypeUploadDebtChase:
		return "DEBT_CHASE"
	case ReportTypeUploadDeputySchedule:
		return "DEPUTY_SCHEDULE"
	case ReportTypeUploadDirectDebitsCollections:
		return "DIRECT_DEBITS_COLLECTIONS"
	case ReportTypeUploadMisappliedPayments:
		return "MISAPPLIED_PAYMENTS"
	default:
		return ""
	}
}

func ParseReportUploadType(s string) ReportUploadType {
	value, ok := reportTypeUploadMap[s]
	if !ok {
		return ReportUploadType(0)
	}
	return value
}

func (u ReportUploadType) Valid() bool {
	return u != ReportTypeUploadUnknown
}

func (u ReportUploadType) IsPayment() bool {
	for _, t := range reportUploadPaymentTypes {
		if u == t {
			return true
		}
	}
	return false
}

func (u ReportUploadType) IsReversal() bool {
	for _, t := range reportUploadReversalTypes {
		if u == t {
			return true
		}
	}
	return false
}

func (u ReportUploadType) HasHeader() bool {
	for _, t := range reportUploadNoHeaderTypes {
		if u == t {
			return false
		}
	}
	return true
}

func (u ReportUploadType) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Key())
}

func (u *ReportUploadType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*u = ParseReportUploadType(s)
	return nil
}
