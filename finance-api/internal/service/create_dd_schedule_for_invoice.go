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

	err = s.ReapplyCredit(ctx, clientID, nil)
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

	var (
		dateReceived pgtype.Date
		balance      pgtype.Int4
		client       pgtype.Int4
	)
	_ = dateReceived.Scan(date)
	_ = store.ToInt4(&balance, clientBalance)
	_ = store.ToInt4(&client, clientID)

	exists, _ := s.store.CheckPendingCollection(ctx, store.CheckPendingCollectionParams{
		DateCollected: dateReceived,
		Amount:        balance,
		ClientID:      client,
	})

	return exists
}
