package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"log"
	"strconv"
	"strings"
	"time"
)

type BadRequests struct {
	Reasons []string `json:"reasons"`
}

func (b BadRequests) Error() string {
	return fmt.Sprintf("bad requests: %s", strings.Join(b.Reasons, ", "))
}

func (s *Service) AddManualInvoice(id int, data shared.AddManualInvoice) error {
	var validationsErrors []string

	validInvoiceTypesRaisedDateInPast := []string{
		shared.InvoiceTypeAD.Key(),
		shared.InvoiceTypeSF.Key(),
		shared.InvoiceTypeSE.Key(),
		shared.InvoiceTypeSO.Key(),
	}

	if contains(validInvoiceTypesRaisedDateInPast, data.InvoiceType) {
		if !data.RaisedDate.Time.Before(time.Now()) {
			validationsErrors = append(validationsErrors, "RaisedDateForAnInvoice")
		}
	}

	isStartDateValid := validateStartDate(data.StartDate, data.EndDate)
	if !isStartDateValid {
		validationsErrors = append(validationsErrors, "StartDate")
	}

	isEndDateValid := validateEndDate(data.StartDate, data.EndDate)
	if !isEndDateValid {
		validationsErrors = append(validationsErrors, "EndDate")
	}

	if len(validationsErrors) != 0 {
		return BadRequests{Reasons: validationsErrors}
	}

	ctx := context.Background()

	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(ctx); !errors.Is(err, sql.ErrTxDone) {
			log.Println("Error rolling back add fee reduction transaction:", err)
		}
	}()

	transaction := s.store.WithTx(tx)
	hasCounterGotInvoiceRefYear, err := transaction.CheckCounterForInvoiceRefYear(ctx, strconv.Itoa(data.StartDate.Time.Year())+"InvoiceNumber")
	if err != nil {
		return err
	}

	var getInvoiceCounterForYear store.Counter
	if hasCounterGotInvoiceRefYear {
		getInvoiceCounterForYear, err = transaction.UpdateCounterForInvoiceRefYear(ctx, strconv.Itoa(data.StartDate.Time.Year())+"InvoiceNumber")
		if err != nil {
			return err
		}
	} else {
		getInvoiceCounterForYear, err = transaction.CreateCounterForInvoiceRefYear(ctx, strconv.Itoa(data.StartDate.Time.Year())+"InvoiceNumber")
		if err != nil {
			return err
		}
	}

	invoiceRef := addLeadingZeros(int(getInvoiceCounterForYear.Counter))

	addManualInvoiceQueryArgs := store.AddManualInvoiceParams{
		PersonID:          pgtype.Int4{Int32: int32(id), Valid: true},
		Feetype:           data.InvoiceType,
		Reference:         data.InvoiceType + invoiceRef + "/" + strconv.Itoa(data.StartDate.Time.Year()%100),
		Startdate:         pgtype.Date{Time: data.StartDate.Time, Valid: true},
		Enddate:           pgtype.Date{Time: data.EndDate.Time, Valid: true},
		Amount:            int32(data.Amount),
		Confirmeddate:     pgtype.Date{Time: time.Now(), Valid: true},
		Batchnumber:       pgtype.Int4{},
		Raiseddate:        pgtype.Date{Time: data.RaisedDate.Time, Valid: true},
		Source:            pgtype.Text{String: "Created manually", Valid: true},
		Scheduledfn14date: pgtype.Date{},
		Cacheddebtamount:  pgtype.Int4{},
		Createddate:       pgtype.Date{Time: time.Now(), Valid: true},
		CreatedbyID:       pgtype.Int4{},
		FeeReductionID:    pgtype.Int4{},
	}

	var invoice store.Invoice
	invoice, err = transaction.AddManualInvoice(ctx, addManualInvoiceQueryArgs)
	if err != nil {
		return err
	}

	if data.SupervisionLevel != "" {
		addInvoiceRangeQueryArgs := store.AddInvoiceRangeParams{
			InvoiceID:        pgtype.Int4{Int32: invoice.ID, Valid: true},
			Supervisionlevel: strings.ToUpper(data.SupervisionLevel),
			Fromdate:         pgtype.Date{Time: data.StartDate.Time, Valid: true},
			Todate:           pgtype.Date{Time: data.EndDate.Time, Valid: true},
			Amount:           invoice.Amount,
		}

		err = transaction.AddInvoiceRange(ctx, addInvoiceRangeQueryArgs)
		if err != nil {
			return err
		}
	}

	feeReductionsRawData, err := transaction.GetFeeReductions(ctx, invoice.FinanceClientID.Int32)
	if err != nil {
		return err
	}

	var feeReductionIds []int32

	for _, fee := range feeReductionsRawData {
		feeReductionIds = append(feeReductionIds, fee.ID)
	}

	AddFeeReductionToInvoiceParams := store.AddFeeReductionToInvoiceParams{
		Column1: feeReductionIds,
		ID:      invoice.ID,
	}

	feeReductionId, _ := transaction.AddFeeReductionToInvoice(ctx, AddFeeReductionToInvoiceParams)

	if feeReductionId != 0 {
		var feeReduction store.GetFeeReductionRow
		feeReduction, err = transaction.GetFeeReduction(ctx, feeReductionId)
		if err != nil {
			return err
		}

		err2 := s.AddLedgerAndAllocations(feeReduction.Type, feeReduction.ID, feeReduction.FinanceClientID.Int32, invoice, transaction, ctx)
		if err2 != nil {
			return err2
		}
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return commitErr
	}

	return nil
}

func (s *Service) AddLedgerAndAllocations(feeReductionFeeType string, feeReductionId int32, feeReductionFinanceClientID int32, invoice store.Invoice, transaction *store.Queries, ctx context.Context) error {
	var amount int32 = 0
	switch strings.ToLower(feeReductionFeeType) {
	case "exemption", "hardship":
		amount = invoice.Amount
	case "remission":
		invoiceFeeRangeParams := store.GetInvoiceFeeRangeAmountParams{
			InvoiceID:        pgtype.Int4{Int32: invoice.ID, Valid: true},
			Supervisionlevel: "GENERAL",
		}
		amount, _ = transaction.GetInvoiceFeeRangeAmount(ctx, invoiceFeeRangeParams)
	}
	if amount != 0 {
		ledgerQueryArgs := store.CreateLedgerForFeeReductionParams{
			Method:          feeReductionFeeType + " credit for invoice " + invoice.Reference,
			Amount:          amount,
			Notes:           pgtype.Text{String: "Credit due to manual invoice " + feeReductionFeeType},
			Type:            "CREDIT " + feeReductionFeeType,
			FinanceClientID: pgtype.Int4{Int32: feeReductionFinanceClientID, Valid: true},
			FeeReductionID:  pgtype.Int4{Int32: feeReductionId, Valid: true},
			//TODO make sure we have correct createdby ID in ticket PFS-88
			CreatedbyID: pgtype.Int4{Int32: 1},
		}
		var ledger store.Ledger
		ledger, err := transaction.CreateLedgerForFeeReduction(ctx, ledgerQueryArgs)
		if err != nil {
			return err
		}

		queryArgs := store.CreateLedgerAllocationForFeeReductionParams{
			LedgerID:  pgtype.Int4{Int32: ledger.ID, Valid: true},
			InvoiceID: pgtype.Int4{Int32: invoice.ID, Valid: true},
			Amount:    ledger.Amount,
		}

		_, err = transaction.CreateLedgerAllocationForFeeReduction(ctx, queryArgs)
		if err != nil {
			return err
		}
	}
	return nil
}

func addLeadingZeros(number int) string {
	strNumber := strconv.Itoa(number)
	paddedNumber := fmt.Sprintf("%0*s", 6, strNumber)

	if len(paddedNumber) > 6 {
		paddedNumber = paddedNumber[len(paddedNumber)-6:]
	}

	return paddedNumber
}

func validateEndDate(startDate *shared.Date, endDate *shared.Date) bool {
	if endDate.Time.Before(startDate.Time) {
		return false
	}

	if !isSameFinancialYear(startDate, endDate) {
		return false
	}

	return true
}

func validateStartDate(startDate *shared.Date, endDate *shared.Date) bool {
	if startDate.Time.After(endDate.Time) {
		return false
	}

	if !isSameFinancialYear(startDate, endDate) {
		return false
	}

	return true
}

func isSameFinancialYear(startDate *shared.Date, endDate *shared.Date) bool {
	var financialYearOneStartYear int
	if startDate.Time.Month() < time.April {
		financialYearOneStartYear = startDate.Time.Year() - 1
	} else {
		financialYearOneStartYear = startDate.Time.Year()
	}
	financialYearOneStart := time.Date(financialYearOneStartYear, time.April, 1, 0, 0, 0, 0, time.UTC)
	financialYearOneEnd := time.Date(financialYearOneStartYear+1, time.March, 31, 23, 59, 59, 999999999, time.UTC)

	var financialYearTwoStartYear int
	if endDate.Time.Month() < time.April {
		financialYearTwoStartYear = endDate.Time.Year() - 1
	} else {
		financialYearTwoStartYear = endDate.Time.Year()
	}
	financialYearTwoStart := time.Date(financialYearTwoStartYear, time.April, 1, 0, 0, 0, 0, time.UTC)
	financialYearTwoEnd := time.Date(financialYearTwoStartYear+1, time.March, 31, 23, 59, 59, 999999999, time.UTC)

	return financialYearOneStart == financialYearTwoStart && financialYearOneEnd == financialYearTwoEnd
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
