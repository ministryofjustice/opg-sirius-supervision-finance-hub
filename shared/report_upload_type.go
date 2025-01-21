package shared

import (
	"encoding/json"
	"fmt"
	"time"
)

var ReportUploadTypes = []ReportUploadType{
	ReportTypeUploadPaymentsMOTOCard,
	ReportTypeUploadPaymentsOnlineCard,
	ReportTypeUploadPaymentsOPGBACS,
	ReportTypeUploadPaymentsSupervisionBACS,
	ReportTypeUploadDebtChase,
	ReportTypeUploadDeputySchedule,
}

type ReportUploadType int

const (
	ReportTypeUploadUnknown ReportUploadType = iota
	ReportTypeUploadPaymentsMOTOCard
	ReportTypeUploadPaymentsOnlineCard
	ReportTypeUploadPaymentsOPGBACS
	ReportTypeUploadPaymentsSupervisionBACS
	ReportTypeUploadDebtChase
	ReportTypeUploadDeputySchedule
)

var reportTypeUploadMap = map[string]ReportUploadType{
	"PAYMENTS_MOTO_CARD":        ReportTypeUploadPaymentsMOTOCard,
	"PAYMENTS_ONLINE_CARD":      ReportTypeUploadPaymentsOnlineCard,
	"PAYMENTS_OPG_BACS":         ReportTypeUploadPaymentsOPGBACS,
	"PAYMENTS_SUPERVISION_BACS": ReportTypeUploadPaymentsSupervisionBACS,
	"DEBT_CHASE":                ReportTypeUploadDebtChase,
	"DEPUTY_SCHEDULE":           ReportTypeUploadDeputySchedule,
}

func (i ReportUploadType) String() string {
	return i.Key()
}

func (i ReportUploadType) Translation() string {
	switch i {
	case ReportTypeUploadPaymentsMOTOCard:
		return "Payments - MOTO card"
	case ReportTypeUploadPaymentsOnlineCard:
		return "Payments - Online card"
	case ReportTypeUploadPaymentsOPGBACS:
		return "Payments - OPG BACS"
	case ReportTypeUploadPaymentsSupervisionBACS:
		return "Payments - Supervision BACS"
	case ReportTypeUploadDebtChase:
		return "Debt chase"
	case ReportTypeUploadDeputySchedule:
		return "Deputy schedule"
	default:
		return ""
	}
}

func (i ReportUploadType) Key() string {
	switch i {
	case ReportTypeUploadPaymentsMOTOCard:
		return "PAYMENTS_MOTO_CARD"
	case ReportTypeUploadPaymentsOnlineCard:
		return "PAYMENTS_ONLINE_CARD"
	case ReportTypeUploadPaymentsOPGBACS:
		return "PAYMENTS_OPG_BACS"
	case ReportTypeUploadPaymentsSupervisionBACS:
		return "PAYMENTS_SUPERVISION_BACS"
	case ReportTypeUploadDebtChase:
		return "DEBT_CHASE"
	case ReportTypeUploadDeputySchedule:
		return "DEPUTY_SCHEDULE"
	default:
		return ""
	}
}

func (i ReportUploadType) CSVHeaders() []string {
	switch i {
	case ReportTypeUploadPaymentsMOTOCard, ReportTypeUploadPaymentsOnlineCard:
		return []string{"Ordercode", "Date", "Amount"}
	case ReportTypeUploadDeputySchedule:
		return []string{"Deputy number", "Deputy name", "Case number", "Client forename", "Client surname", "Do not invoice", "Total outstanding"}
	case ReportTypeUploadDebtChase:
		return []string{"Client_no", "Deputy_name", "Total_debt"}
	case ReportTypeUploadPaymentsSupervisionBACS:
		return []string{"Line", "Type", "Code", "Number", "Transaction Date", "Value Date", "Amount", "Amount Reconciled", "Charges", "Status", "Desc Flex", "Consolidated line"}
	case ReportTypeUploadPaymentsOPGBACS:
		return []string{"Line", "Type", "Code", "Number", "Transaction Date", "Value Date", "Amount", "Amount Reconciled", "Charges", "Status", "Desc Flex"}
	}

	return []string{"Unknown report type"}
}

func (i ReportUploadType) Filename(date string) (string, error) {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", err
	}

	switch i {
	case ReportTypeUploadPaymentsMOTOCard:
		return fmt.Sprintf("feemoto_%snormal.csv", parsedDate.Format("02012006")), nil
	case ReportTypeUploadPaymentsOnlineCard:
		return fmt.Sprintf("feemoto_%smlpayments.csv", parsedDate.Format("02012006")), nil
	case ReportTypeUploadPaymentsSupervisionBACS:
		return fmt.Sprintf("feebacs_%s_new_acc.csv", parsedDate.Format("02012006")), nil
	case ReportTypeUploadPaymentsOPGBACS:
		return fmt.Sprintf("feebacs_%s.csv", parsedDate.Format("02012006")), nil
	default:
		return "", nil
	}
}

func ParseReportUploadType(s string) ReportUploadType {
	value, ok := reportTypeUploadMap[s]
	if !ok {
		return ReportUploadType(0)
	}
	return value
}

func (i ReportUploadType) Valid() bool {
	return i != ReportTypeUploadUnknown
}

func (i ReportUploadType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Key())
}

func (i *ReportUploadType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*i = ParseReportUploadType(s)
	return nil
}
