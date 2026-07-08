package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

const billingDay = 24

type PendingCollection struct {
	Amount         int32
	CollectionDate time.Time
}

func (s *Service) CreateDirectDebitSchedule(ctx context.Context, details shared.InvoiceCreatedEvent) error {
	logger := s.Logger(ctx)

	if !s.env.AllpayEnabled {
		logger.Info(fmt.Sprintf("skipping Direct Debit schedule creation for client id %d as Allpay is disabled in this environment", details.ClientID))
		return nil
	}

	if !details.InvoiceType.IsDirectDebitInvoice() {
		return nil
	}

	client, err := s.store.GetClientById(ctx, details.ClientID)
	if err != nil {
		return err
	}
	if client.PaymentMethod != shared.PaymentMethodDirectDebit.Key() {
		return nil
	}

	if s.pendingScheduleExists(ctx, details.ClientID) {
		logger.Info(fmt.Sprintf("skipping Direct Debit schedule creation for client %d as a schedule already exists", details.ClientID))
		return nil
	}

	schedule, err := s.generateScheduleData(ctx, details.ClientID)
	if err != nil {
		return err
	}

	if schedule.Amount < 1 {
		logger.Info(fmt.Sprintf("skipping Direct Debit schedule creation for client %d as there is no balance outstanding", details.ClientID))
		return nil
	}

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var collectionDate pgtype.Date
	_ = collectionDate.Scan(schedule.CollectionDate)

	err = tx.CreatePendingCollection(ctx, store.CreatePendingCollectionParams{
		ClientID:       details.ClientID,
		CollectionDate: collectionDate,
		Amount:         schedule.Amount,
		CreatedBy:      ctx.(auth.Context).User.ID,
	})
	if err != nil {
		logger.Error("failed to create pending collection for Direct Debit schedule, aborting", "error", err)
		return err
	}

	err = s.allpay.CreateSchedule(ctx, &allpay.CreateScheduleInput{
		ClientDetails: allpay.ClientDetails{
			ClientReference: client.CourtRef,
			Surname:         client.Surname,
		},
		Date:   schedule.CollectionDate,
		Amount: schedule.Amount,
	})
	if err != nil {
		var ve allpay.ErrorValidation
		if errors.As(err, &ve) {
			logger.Error("validation errors returned from allpay", "errors", ve.Messages)
		}
		dispatchErr := s.dispatch.DirectDebitScheduleFailed(ctx, event.DirectDebitScheduleFailed{
			ClientID: int(details.ClientID),
		})
		if dispatchErr != nil {
			return dispatchErr
		}
		return apierror.BadRequestError("Allpay", "Failed", err)
	}
	return tx.Commit(ctx)
}

func (s *Service) pendingScheduleExists(ctx context.Context, clientID int32) bool {
	clientBalance, _ := s.store.GetPendingOutstandingBalance(ctx, clientID)
	date, _ := s.calculateScheduleCollectionDate(ctx)

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

// generateScheduleData creates a schedule object by checking outstanding balance and calculating collection date.
// Returns an empty PendingCollection if there is no balance.
func (s *Service) generateScheduleData(ctx context.Context, clientID int32) (PendingCollection, error) {
	balance, err := s.store.GetPendingOutstandingBalance(ctx, clientID)
	if err != nil {
		s.Logger(ctx).Error("failed to fetch outstanding balance", "client_id", clientID, "error", err)
		return PendingCollection{}, err
	}

	if balance < 1 {
		return PendingCollection{}, nil
	}

	date, err := s.calculateScheduleCollectionDate(ctx)
	if err != nil {
		s.Logger(ctx).Error("failed to calculate collection date", "client_id", clientID, "error", err)
		return PendingCollection{}, err
	}

	return PendingCollection{
		Amount:         balance,
		CollectionDate: date,
	}, nil
}

func (s *Service) calculateScheduleCollectionDate(ctx context.Context) (time.Time, error) {
	date, err := s.govUK.AddWorkingDays(ctx, time.Now().UTC(), 14)
	if err != nil {
		return date, err
	}
	return s.govUK.NextWorkingDayOnOrAfterX(ctx, date, billingDay)
}
