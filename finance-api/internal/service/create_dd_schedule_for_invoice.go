package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
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

	if s.pendingScheduleExists(ctx, clientID) {
		logger.Info(fmt.Sprintf("skipping direct debit schedule creation for invoice %d as a schedule already exists", data.InvoiceId))
		return nil
	}

	pendingCollection, err := s.CreateDirectDebitSchedule(ctx, clientID, shared.CreateSchedule{AllPayCustomer: data.AllPayCustomer})
	if err != nil {
		return err
	}

	var pc PendingCollection

	if pendingCollection != pc {
		if err := s.SendDirectDebitCollectionEvent(ctx, clientID, pendingCollection); err != nil {
			logger.Error("Sending direct-debit-collection event in CreateDirectDebitScheduleForInvoice failed", "err", err)
			return err
		}
	}

	return nil
}

func (s *Service) pendingScheduleExists(ctx context.Context, clientID int32) bool {
	clientBalance, _ := s.store.GetPendingOutstandingBalance(ctx, clientID)
	date, _ := s.CalculateScheduleCollectionDate(ctx)

	exists, _ := s.store.CheckPendingCollection(ctx, store.CheckPendingCollectionParams{
		DateCollected: pgtype.Date{Time: date, Valid: true},
		Amount:        pgtype.Int4{Int32: clientBalance, Valid: true},
		ClientID:      pgtype.Int4{Int32: clientID, Valid: true},
	})

	return exists != 0
}
