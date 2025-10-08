package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"time"
)

func (s *Service) CreateDirectDebitScheduleForInvoice(ctx context.Context, clientID int32, data shared.CreateScheduleForInvoice) error {
	var err error

	logger := s.Logger(ctx)

	err = s.reapplyCreditTx(ctx, clientID)
	if err != nil {
		return err
	}

	balance, err := s.store.GetInvoiceBalanceDetails(ctx, data.InvoiceId)
	if err != nil {
		return err
	}

	if balance.Outstanding < 1 {
		logger.Info(fmt.Sprintf("skipping direct debit schedule creation for invoice %d as there is no balance outstanding", data.InvoiceId), "balance", balance)
		return nil
	}

	// Check if schedule has already been made for this invoice
	if s.pendingScheduleExists(ctx, clientID) {
		logger.Info(fmt.Sprintf("skipping direct debit schedule creation for invoice %d as a schedule already exists", data.InvoiceId))
		return nil
	}

	err = s.CreateDirectDebitSchedule(ctx, clientID, shared.CreateSchedule{AllPayCustomer: data.AllPayCustomer})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) pendingScheduleExists(ctx context.Context, clientID int32) bool {
	clientBalance, _ := s.store.GetPendingOutstandingBalance(ctx, clientID)
	date, _ := s.govUK.AddWorkingDays(ctx, time.Now().UTC(), 14)

	exists, _ := s.store.CheckPendingCollection(ctx, store.CheckPendingCollectionParams{
		DateCollected:   pgtype.Date{Time: date, Valid: true},
		Amount:          clientBalance,
		FinanceClientID: pgtype.Int4{Int32: clientID, Valid: true},
	})

	return exists != 0
}
