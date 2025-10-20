package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) CancelDirectDebitMandate(ctx context.Context, clientID int32, cancelMandate shared.CancelMandate) error {
	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// update payment method first, in case this fails
	err = tx.UpdatePaymentMethod(ctx, store.UpdatePaymentMethodParams{
		PaymentMethod: shared.PaymentMethodDemanded.Key(),
		ClientID:      clientID,
	})
	if err != nil {
		return err
	}

	collections, err := tx.GetPendingCollections(ctx, clientID)
	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error fetching pending collections for mandate cancellation, rolling back payment method change for client : %d", clientID), slog.String("err", err.Error()))
		return err
	}

	closureDate, err := s.calculateClosureDate(ctx, collections)
	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error calculating closure date for mandate, rolling back payment method change for client : %d", clientID), slog.String("err", err.Error()))
		return err
	}

	// update allpay
	clientDetails := allpay.ClientDetails{
		ClientReference: cancelMandate.ClientReference,
		Surname:         cancelMandate.Surname,
	}

	err = s.allpay.CancelMandate(ctx, &allpay.CancelMandateRequest{
		ClosureDate:   closureDate,
		ClientDetails: clientDetails,
	})

	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error cancelling mandate with allpay, rolling back payment method change for client : %d", clientID), slog.String("err", err.Error()))
		return err
	}

	err = s.dispatch.PaymentMethodChanged(ctx, event.PaymentMethod{
		ClientID:      int(clientID),
		PaymentMethod: shared.PaymentMethodDemanded,
	})
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	// remove schedules after committing the transaction, as the mandate has already been cancelled in Allpay, and we are unable to roll that back
	return s.cancelPendingCollections(ctx, closureDate, clientDetails, collections)
}

/**
 * calculateClosureDate returns the closure date for the mandate. As BACS takes three working days to process a Direct Debit
 * collection, we have no certainty for whether a pending collection in that period will be collected. To avoid a situation
 * where we record a payment that isn't processed, or fail to process a collection that is collected, this function finds
 * the pending collections that fall within those three working days and sets the closure date to the next working day.
 */
func (s *Service) calculateClosureDate(ctx context.Context, collections []store.GetPendingCollectionsRow) (time.Time, error) {
	closureDate := time.Now().Truncate(24 * time.Hour)
	bacsDate, err := s.govUK.AddWorkingDays(ctx, closureDate, 3)
	if err != nil {
		return time.Time{}, err
	}
	for _, pc := range collections {
		if pc.CollectionDate.Time.Truncate(24 * time.Hour).After(bacsDate) {
			break
		}
		closureDate = pc.CollectionDate.Time.AddDate(0, 0, 1)
	}
	return s.govUK.NextWorkingDayOnOrAfterX(ctx, closureDate, closureDate.Day()) // will return the closure date if it is a working day
}

func (s *Service) cancelPendingCollections(ctx context.Context, closureDate time.Time, clientDetails allpay.ClientDetails, collections []store.GetPendingCollectionsRow) error {
	for _, pc := range collections {
		if pc.CollectionDate.Time.After(closureDate) {
			err := s.allpay.RemoveScheduledPayment(ctx, &allpay.RemoveScheduledPaymentRequest{
				CollectionDate: pc.CollectionDate.Time,
				Amount:         pc.Amount,
				ClientDetails:  clientDetails,
			})
			if err != nil {
				return err
			}
			err = s.store.CancelPendingCollection(ctx, pc.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
