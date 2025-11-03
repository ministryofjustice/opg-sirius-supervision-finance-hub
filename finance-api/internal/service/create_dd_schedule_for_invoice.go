package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) CreateDirectDebitScheduleForInvoice(ctx context.Context, clientID int32) error {
	var err error

	logger := s.Logger(ctx)

	client, err := s.store.GetClientById(ctx, clientID)
	if err != nil {
		return err
	}
	if client.PaymentMethod != shared.PaymentMethodDirectDebit.Key() {
		return nil
	}

	if s.pendingScheduleExists(ctx, clientID) {
		logger.Info(fmt.Sprintf("skipping direct debit schedule creation for client %d as a schedule already exists", clientID))
		return nil
	}

	allPayCustomer := shared.AllPayCustomer{
		Surname:         client.Surname.String,
		ClientReference: client.CourtRef.String,
	}
	_, err = s.CreateDirectDebitSchedule(ctx, clientID, shared.CreateSchedule{AllPayCustomer: allPayCustomer})
	if err != nil {
		return err
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
