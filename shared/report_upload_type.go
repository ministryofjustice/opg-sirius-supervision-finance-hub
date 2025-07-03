package shared

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"
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
	ReportTypeUploadDuplicatedPayments,
	ReportTypeUploadBouncedCheque,
	ReportTypeUploadFailedDirectDebitCollections,
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
	ReportTypeUploadDuplicatedPayments
	ReportTypeUploadBouncedCheque
	ReportTypeUploadFulfilledRefunds
	ReportTypeUploadFailedDirectDebitCollections
)

var reportTypeUploadMap = map[string]ReportUploadType{
	"PAYMENTS_MOTO_CARD":               ReportTypeUploadPaymentsMOTOCard,
	"PAYMENTS_ONLINE_CARD":             ReportTypeUploadPaymentsOnlineCard,
	"PAYMENTS_OPG_BACS":                ReportTypeUploadPaymentsOPGBACS,
	"PAYMENTS_SUPERVISION_BACS":        ReportTypeUploadPaymentsSupervisionBACS,
	"PAYMENTS_SUPERVISION_CHEQUE":      ReportTypeUploadPaymentsSupervisionCheque,
	"DEBT_CHASE":                       ReportTypeUploadDebtChase,
	"DEPUTY_SCHEDULE":                  ReportTypeUploadDeputySchedule,
	"DIRECT_DEBITS_COLLECTIONS":        ReportTypeUploadDirectDebitsCollections,
	"MISAPPLIED_PAYMENTS":              ReportTypeUploadMisappliedPayments,
	"DUPLICATED_PAYMENTS":              ReportTypeUploadDuplicatedPayments,
	"BOUNCED_CHEQUE":                   ReportTypeUploadBouncedCheque,
	"FULFILLED_REFUNDS":                ReportTypeUploadFulfilledRefunds,
	"FAILED_DIRECT_DEBITS_COLLECTIONS": ReportTypeUploadFailedDirectDebitCollections,
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
		return "Direct Debits Collections"
	case ReportTypeUploadMisappliedPayments:
		return "Payment Reversals - Misapplied payments"
	case ReportTypeUploadDuplicatedPayments:
		return "Payment Reversals - Duplicated payments"
	case ReportTypeUploadBouncedCheque:
		return "Payment Reversals - Bounced cheque"
	case ReportTypeUploadFulfilledRefunds:
		return "Fulfilled refunds"
	case ReportTypeUploadFailedDirectDebitCollections:
		return "Payment Reversals - Failed direct debit collections"
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
	case ReportTypeUploadDuplicatedPayments:
		return "DUPLICATED_PAYMENTS"
	case ReportTypeUploadBouncedCheque:
		return "BOUNCED_CHEQUE"
	case ReportTypeUploadFulfilledRefunds:
		return "FULFILLED_REFUNDS"
	case ReportTypeUploadFailedDirectDebitCollections:
		return "FAILED_DIRECT_DEBITS_COLLECTIONS"
	default:
		return ""
	}
}

func (u ReportUploadType) CSVHeaders() []string {
	switch u {
	case ReportTypeUploadPaymentsMOTOCard, ReportTypeUploadPaymentsOnlineCard:
		return []string{"Ordercode", "Date", "Amount"}
	case ReportTypeUploadPaymentsSupervisionBACS:
		return []string{"Line", "Type", "Code", "Number", "Transaction Date", "Value Date", "Amount", "Amount Reconciled", "Charges", "Status", "Desc Flex", "Consolidated line"}
	case ReportTypeUploadPaymentsOPGBACS:
		return []string{"Line", "Type", "Code", "Number", "Transaction Date", "Value Date", "Amount", "Amount Reconciled", "Charges", "Status", "Desc Flex"}
	case ReportTypeUploadPaymentsSupervisionCheque:
		return []string{"Case number (confirmed on Sirius)", "Cheque number", "Cheque Value (£)", "Comments", "Date in Bank"}
	case ReportTypeUploadDeputySchedule:
		return []string{"Deputy number", "Deputy name", "Case number", "Client forename", "Client surname", "Do not invoice", "Total outstanding"}
	case ReportTypeUploadDebtChase:
		return []string{"Case_no", "Client_no", "Client_title", "Client_forename", "Client_surname", "Do_not_chase", "Payment_method", "Deputy_type", "Deputy_no", "Airmail", "Deputy_title", "Deputy_Welsh", "Deputy_Large_Print", "Deputy_name", "Email", "Address1", "Address2", "Address3", "City_Town", "County", "Postcode", "Total_debt", "Invoice1", "Amount1"}
	case ReportTypeUploadMisappliedPayments:
		return []string{"Payment type", "Current (errored) court reference", "New (correct) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"}
	case ReportTypeUploadDuplicatedPayments:
		return []string{"Payment type", "Current (errored) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"}
	case ReportTypeUploadBouncedCheque:
		return []string{"Court reference", "Bank date", "Received date", "Amount", "PIS number"}
	case ReportTypeUploadFulfilledRefunds:
		return []string{"Court reference", "Amount", "Bank account name", "Bank account number", "Bank account sort code", "Created by", "Approved by"}
	case ReportTypeUploadFailedDirectDebitCollections:
		return []string{"Court reference", "Bank date", "Received date", "Amount"}
	}

	return []string{"Unknown report type"}
}

func (u ReportUploadType) Filename(date string) (string, error) {
	switch u {
	case ReportTypeUploadMisappliedPayments:
		return "misappliedpayments.csv", nil
	case ReportTypeUploadBouncedCheque:
		return "bouncedcheque.csv", nil
	case ReportTypeUploadDuplicatedPayments:
		return "duplicatedpayments.csv", nil
	case ReportTypeUploadDebtChase:
		return "debt_FeeChase_*.csv", nil
	case ReportTypeUploadFailedDirectDebitCollections:
		return "faileddirectdebitcollections.csv", nil
	}

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", err
	}

	switch u {
	case ReportTypeUploadPaymentsMOTOCard:
		return fmt.Sprintf("feemoto_%snormal.csv", parsedDate.Format("02012006")), nil
	case ReportTypeUploadPaymentsOnlineCard:
		return fmt.Sprintf("feemoto_%smlpayments.csv", parsedDate.Format("02012006")), nil
	case ReportTypeUploadPaymentsSupervisionBACS:
		return fmt.Sprintf("feebacs_%s_new_acc.csv", parsedDate.Format("02012006")), nil
	case ReportTypeUploadPaymentsOPGBACS:
		return fmt.Sprintf("feebacs_%s.csv", parsedDate.Format("02012006")), nil
	case ReportTypeUploadPaymentsSupervisionCheque:
		return fmt.Sprintf("supervisioncheques_%s.csv", parsedDate.Format("02012006")), nil
	case ReportTypeUploadDirectDebitsCollections:
		return fmt.Sprintf("directdebitscollections_%s.csv", parsedDate.Format("02012006")), nil
	default:
		return "", nil
	}
}

func ParseUploadType(s string) ReportUploadType {
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
	return slices.Contains(reportUploadPaymentTypes, u)
}

func (u ReportUploadType) IsReversal() bool {
	return slices.Contains(reportUploadReversalTypes, u)
}

func (u ReportUploadType) IsRefund() bool {
	return u == ReportTypeUploadFulfilledRefunds
}

func (u ReportUploadType) IsDirectUpload() bool {
	return u == ReportTypeUploadDebtChase
}

func (u ReportUploadType) HasHeader() bool {
	return !slices.Contains(reportUploadNoHeaderTypes, u)
}

func (u ReportUploadType) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Key())
}

func (u *ReportUploadType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*u = ParseUploadType(s)
	return nil
}
