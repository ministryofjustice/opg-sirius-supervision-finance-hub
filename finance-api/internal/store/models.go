// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0

package store

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Account struct {
	Code                   int64
	AccountCodeDescription string
	CostCentre             pgtype.Int4
}

type Assignee struct {
	ID      int32
	Name    pgtype.Text
	Surname pgtype.Text
}

type BillingPeriod struct {
	ID              int32
	FinanceClientID pgtype.Int4
	OrderID         pgtype.Int4
	StartDate       pgtype.Date
	EndDate         pgtype.Date
}

type Case struct {
	ID          int32
	ClientID    pgtype.Int4
	Orderstatus pgtype.Text
}

type CostCentre struct {
	Code                  int32
	CostCentreDescription string
}

type Counter struct {
	ID      int32
	Key     string
	Counter int32
}

type FeeReduction struct {
	ID              int32
	FinanceClientID pgtype.Int4
	// (DC2Type:refdata)
	Type string
	// (DC2Type:refdata)
	Evidencetype       pgtype.Text
	Startdate          pgtype.Date
	Enddate            pgtype.Date
	Notes              string
	Deleted            bool
	Datereceived       pgtype.Date
	CreatedAt          pgtype.Timestamp
	CreatedBy          pgtype.Int4
	CancelledAt        pgtype.Timestamp
	CancelledBy        pgtype.Int4
	CancellationReason pgtype.Text
}

type FinanceClient struct {
	ID        int32
	ClientID  int32
	SopNumber string
	// (DC2Type:refdata)
	PaymentMethod string
	Batchnumber   pgtype.Int4
	CourtRef      pgtype.Text
}

type Invoice struct {
	ID              int32
	PersonID        pgtype.Int4
	FinanceClientID pgtype.Int4
	Feetype         string
	Reference       string
	Startdate       pgtype.Date
	Enddate         pgtype.Date
	// (DC2Type:money)
	Amount int32
	// (DC2Type:refdata)
	Supervisionlevel  pgtype.Text
	Confirmeddate     pgtype.Date
	Batchnumber       pgtype.Int4
	Raiseddate        pgtype.Date
	Source            pgtype.Text
	Scheduledfn14date pgtype.Date
	// (DC2Type:money)
	Cacheddebtamount pgtype.Int4
	CreatedAt        pgtype.Timestamp
	CreatedBy        pgtype.Int4
}

type InvoiceAdjustment struct {
	ID              int32
	FinanceClientID int32
	InvoiceID       int32
	RaisedDate      pgtype.Date
	AdjustmentType  string
	Amount          int32
	Notes           string
	Status          string
	CreatedAt       pgtype.Timestamp
	CreatedBy       int32
	UpdatedAt       pgtype.Timestamp
	UpdatedBy       pgtype.Int4
	LedgerID        pgtype.Int4
}

type InvoiceEmailStatus struct {
	ID        int32
	InvoiceID pgtype.Int4
	// (DC2Type:refdata)
	Status string
	// (DC2Type:refdata)
	Templateid  string
	Createddate pgtype.Date
}

type InvoiceFeeRange struct {
	ID        int32
	InvoiceID pgtype.Int4
	// (DC2Type:refdata)
	Supervisionlevel string
	Fromdate         pgtype.Date
	Todate           pgtype.Date
	// (DC2Type:money)
	Amount int32
}

type Ledger struct {
	ID        int32
	Reference string
	Datetime  pgtype.Timestamp
	Method    string
	// (DC2Type:money)
	Amount int32
	Notes  pgtype.Text
	// (DC2Type:refdata)
	Type string
	// (DC2Type:refdata)
	Status          string
	FinanceClientID pgtype.Int4
	ParentID        pgtype.Int4
	FeeReductionID  pgtype.Int4
	Confirmeddate   pgtype.Date
	Bankdate        pgtype.Date
	Batchnumber     pgtype.Int4
	// (DC2Type:refdata)
	Bankaccount pgtype.Text
	Source      pgtype.Text
	Line        pgtype.Int4
	CreatedAt   pgtype.Timestamp
	CreatedBy   pgtype.Int4
}

type LedgerAllocation struct {
	ID        int32
	LedgerID  pgtype.Int4
	InvoiceID pgtype.Int4
	Datetime  pgtype.Timestamp
	// (DC2Type:money)
	Amount int32
	// (DC2Type:refdata)
	Status        string
	Reference     pgtype.Text
	Notes         pgtype.Text
	Allocateddate pgtype.Date
	Batchnumber   pgtype.Int4
	Source        pgtype.Text
}

type Person struct {
	ID            int32
	Firstname     pgtype.Text
	Surname       pgtype.Text
	Caserecnumber pgtype.Text
	FeepayerID    pgtype.Int4
	Deputytype    pgtype.Text
}

type Property struct {
	ID    int32
	Key   string
	Value string
}

type Rate struct {
	ID        int32
	Type      string
	Startdate pgtype.Date
	Enddate   pgtype.Date
	// (DC2Type:money)
	Amount int32
}

type Report struct {
	ID          int32
	Batchnumber int32
	// (DC2Type:refdata)
	Type        string
	Datetime    pgtype.Timestamp
	Count       int32
	Invoicedate pgtype.Timestamp
	// (DC2Type:money)
	Totalamount           pgtype.Int4
	Firstinvoicereference pgtype.Text
	Lastinvoicereference  pgtype.Text
	CreatedbyuserID       pgtype.Int4
}

type TransactionType struct {
	ID               int32
	FeeType          string
	SupervisionLevel string
	LedgerType       pgtype.Text
	AccountCode      pgtype.Int8
	Description      pgtype.Text
	LineDescription  pgtype.Text
	IsReceipt        pgtype.Bool
}

type TransactionTypeUpdate struct {
	TransactionTypeID     int32
	LineDescriptionUpdate pgtype.Text
	IsReceiptUpdate       pgtype.Bool
}
