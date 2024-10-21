package service

import (
	"context"
	"encoding/csv"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"strconv"
	"strings"
	"time"
)

func (s *Service) ProcessFinanceAdminUpload(ctx context.Context, bucketName string, key string) error {
	file, err := s.filestorage.GetFile(ctx, bucketName, key)

	if err != nil {
		return err
	}

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		return err
	}

	for index, record := range records {
		if index != 0 {
			err := s.processMotoCardPaymentsUploadLine(ctx, record)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func parseAmount(amount string) (int32, error) {
	index := strings.Index(amount, ".")

	if index != -1 && len(amount)-index == 2 {
		amount = amount + "0"
	} else if index == -1 {
		amount = amount + "00"
	}

	intAmount, err := strconv.Atoi(strings.Replace(amount, ".", "", 1))
	return int32(intAmount), err
}

func (s *Service) processMotoCardPaymentsUploadLine(ctx context.Context, record []string) error {
	ctx, cancelTx := context.WithCancel(ctx)
	defer cancelTx()

	courtReference := strings.SplitN(record[0], "-", -1)[0]

	amount, err := parseAmount(record[2])
	if err != nil {
		return err
	}

	parsedDate, err := time.Parse("2006-01-02 15:04:05", record[1])
	if err != nil {
		return err
	}

	if amount == 0 {
		return nil
	}

	// check if payment has already been made
	ledgerId, _ := s.store.GetLedgerForPayment(ctx, store.GetLedgerForPaymentParams{
		CourtRef: pgtype.Text{String: courtReference, Valid: true},
		Amount:   amount,
		Type:     "MOTO card payment",
		Datetime: pgtype.Timestamp{Time: parsedDate, Valid: true},
	})

	if ledgerId != 0 {
		return nil
	}

	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}

	transaction := s.store.WithTx(tx)

	ledgerId, err = transaction.CreateLedgerForCourtRef(ctx, store.CreateLedgerForCourtRefParams{
		CourtRef:  pgtype.Text{String: courtReference, Valid: true},
		Amount:    amount,
		Type:      "MOTO card payment",
		Status:    "CONFIRMED",
		CreatedBy: pgtype.Int4{Int32: 1, Valid: true},
		Datetime:  pgtype.Timestamp{Time: parsedDate, Valid: true},
	})

	if err != nil {
		return err
	}

	invoices, err := transaction.GetInvoicesForCourtRef(ctx, pgtype.Text{String: courtReference, Valid: true})
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		allocationAmount := invoice.Outstanding
		if allocationAmount > amount {
			allocationAmount = amount
		}

		err = transaction.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			InvoiceID: pgtype.Int4{Int32: invoice.ID, Valid: true},
			Amount:    allocationAmount,
			Status:    "ALLOCATED",
			LedgerID:  pgtype.Int4{Int32: ledgerId, Valid: true},
		})
		if err != nil {
			return err
		}

		amount -= allocationAmount
	}

	if amount > 0 {
		err = transaction.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			Amount:   amount,
			Status:   "UNAPPLIED",
			LedgerID: pgtype.Int4{Int32: ledgerId, Valid: true},
		})

		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
