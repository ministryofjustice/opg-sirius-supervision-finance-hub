package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"slices"
	"strconv"
	"time"
)

func (s *Service) AddFeeReduction(ctx context.Context, clientId int32, data shared.AddFeeReduction) error {
	overlapParams := store.CountOverlappingFeeReductionParams{
		ClientID:   clientId,
		Overlaps:   calculateFeeReductionStartDate(data.StartYear),
		Overlaps_2: calculateFeeReductionEndDate(data.StartYear, data.LengthOfAward),
	}

	hasFeeReduction, _ := s.store.CountOverlappingFeeReduction(ctx, overlapParams)
	if hasFeeReduction != 0 {
		return apierror.BadRequest{Reason: "overlap"}
	}

	var (
		dateReceived pgtype.Date
		createdBy    pgtype.Int4
	)
	_ = dateReceived.Scan(data.DateReceived.Time)
	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	feeReductionParams := store.AddFeeReductionParams{
		ClientID:     clientId,
		Type:         data.FeeType.Key(),
		Startdate:    calculateFeeReductionStartDate(data.StartYear),
		Enddate:      calculateFeeReductionEndDate(data.StartYear, data.LengthOfAward),
		Notes:        data.Notes,
		Datereceived: dateReceived,
		CreatedBy:    createdBy,
	}

	ctx, cancelTx := s.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}

	feeReduction, err := tx.AddFeeReduction(ctx, feeReductionParams)
	if err != nil {
		return err
	}

	invoices, err := tx.GetInvoiceBalancesForFeeReductionRange(ctx, feeReduction.ID)
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		amount := calculateFeeReduction(shared.ParseFeeReductionType(feeReduction.Type), invoice.Amount, invoice.Feetype, invoice.GeneralSupervisionFee)
		ledger, allocations := generateLedgerEntries(ctx, addLedgerVars{
			amount:             amount,
			transactionType:    data.FeeType,
			feeReductionId:     feeReduction.ID,
			clientId:           clientId,
			invoiceId:          invoice.ID,
			outstandingBalance: invoice.Outstanding,
		})
		ledgerID, err := tx.CreateLedger(ctx, ledger)
		if err != nil {
			return err
		}

		for _, allocation := range allocations {
			allocation.LedgerID = ledgerID
			err = tx.CreateLedgerAllocation(ctx, allocation)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return s.ReapplyCredit(ctx, clientId)
}

func calculateFeeReduction(feeReductionType shared.FeeReductionType, invoiceTotal int32, invoiceFeeType string, generalSupervisionFee int32) int32 {
	if feeReductionType != shared.FeeReductionTypeRemission {
		return invoiceTotal
	}

	nonGeneralRemissionFeeTypes := []string{"AD", "GA", "GT", "GS"}
	if slices.Contains(nonGeneralRemissionFeeTypes, invoiceFeeType) {
		return invoiceTotal / 2
	}

	return generalSupervisionFee / 2
}

func calculateFeeReductionEndDate(startYear string, lengthOfAward int) (endDate pgtype.Date) {
	startYearTransformed, _ := strconv.Atoi(startYear)
	endDateString := startYearTransformed + lengthOfAward
	feeReductionEndDate, _ := time.Parse("2006-01-02", strconv.Itoa(endDateString)+"-03-31")
	_ = endDate.Scan(feeReductionEndDate)
	return endDate
}

func calculateFeeReductionStartDate(startYear string) (startDate pgtype.Date) {
	feeReductionStartDate, _ := time.Parse("2006-01-02", startYear+"-04-01")
	_ = startDate.Scan(feeReductionStartDate)
	return startDate
}
