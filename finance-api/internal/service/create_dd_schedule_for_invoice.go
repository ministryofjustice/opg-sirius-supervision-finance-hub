package service

import (
	"context"
	"fmt"
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

	// Check if schedule has already been made for this invoice

	err = s.CreateDirectDebitSchedule(ctx, clientID, shared.CreateSchedule{AllPayCustomer: data.AllPayCustomer})
	if err != nil {
		return err
	}

	return nil
}
