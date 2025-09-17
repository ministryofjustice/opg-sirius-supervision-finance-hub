package service

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (s *Service) CreateDirectDebitSchedule(ctx context.Context, clientID int32, data shared.CreateSchedule) (PendingCollection, error) {
	var pc PendingCollection
	logger := s.Logger(ctx)

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return pc, err
	}
	defer tx.Rollback(ctx)

	balance, err := tx.GetPendingOutstandingBalance(ctx, clientID)
	if err != nil {
		logger.Error("failed to create schedule due to error in fetching outstanding balance", "error", err)
		return pc, err
	}
	if balance < 1 {
		logger.Info(fmt.Sprintf("skipping direct debit schedule creation for client %d as there is no balance outstanding", clientID), "balance", balance)
		return pc, nil
	}

	date, err := s.govUK.AddWorkingDays(ctx, time.Now().UTC(), 14)
	if err != nil {
		logger.Error("failed to create schedule due to error in calculating working days", "error", err)
		return pc, err
	}

	date, _ = s.govUK.NextWorkingDayOnOrAfterX(ctx, date, billingDay) // no need to check error here as it would have failed earlier

	var collectionDate pgtype.Date
	_ = collectionDate.Scan(date)

	err = tx.CreatePendingCollection(ctx, store.CreatePendingCollectionParams{
		ClientID:       clientID,
		CollectionDate: collectionDate,
		Amount:         balance,
		CreatedBy:      ctx.(auth.Context).User.ID,
	})
	if err != nil {
		logger.Error("failed to create pending collection for direct debit schedule, aborting", "error", err)
		return pc, err
	}

	err = s.allpay.CreateSchedule(ctx, &allpay.CreateScheduleInput{
		ClientRef: data.ClientReference,
		Surname:   data.Surname,
		Date:      date,
		Amount:    balance,
	})
	if err != nil {
		var ve allpay.ErrorValidation
		if errors.As(err, &ve) {
			// we validate in advance so validation errors from AllPay should never occur
			// if they do, log them so we can investigate
			logger.Error("validation errors returned from allpay", "errors", ve.Messages)
		}
		return pc, err
	}
	pc.Amount = balance
	pc.CollectionDate = collectionDate.Time
	return pc, tx.Commit(ctx)
}
