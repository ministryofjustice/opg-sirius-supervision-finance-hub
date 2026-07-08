package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) CreateDirectDebitMandate(ctx context.Context, clientID int32, createMandate shared.CreateMandate) (PendingCollection, error) {
	bankDetails := createMandate.BankAccount.BankDetails
	err := s.allpay.ModulusCheck(ctx, bankDetails.SortCode, bankDetails.AccountNumber)
	if err != nil {
		return PendingCollection{}, err
	}

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return PendingCollection{}, err
	}
	defer tx.Rollback(ctx)

	var (
		paymentMethod pgtype.Text
		id            pgtype.Int4
		createdBy     pgtype.Int4
	)

	_ = paymentMethod.Scan(shared.PaymentMethodDirectDebit.Key())
	_ = store.ToInt4(&id, clientID)
	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	// update payment method first, in case this fails
	err = tx.SetPaymentMethod(ctx, store.SetPaymentMethodParams{
		PaymentMethod: paymentMethod,
		ClientID:      id,
		CreatedBy:     createdBy,
	})
	if err != nil {
		return PendingCollection{}, err
	}

	mandateRequest := &allpay.CreateMandateRequest{
		Customer: allpay.Customer{
			ClientReference: createMandate.ClientReference,
			Surname:         createMandate.Surname,
			Address: allpay.Address{
				Line1:    createMandate.Address.Line1,
				Town:     createMandate.Address.Town,
				PostCode: createMandate.Address.PostCode,
			},
		},
		BankAccount: struct {
			BankDetails allpay.BankDetails `json:"BankDetails"`
		}{
			BankDetails: allpay.BankDetails{
				AccountName:   bankDetails.AccountName,
				SortCode:      strings.ReplaceAll(bankDetails.SortCode, "-", ""),
				AccountNumber: bankDetails.AccountNumber,
			},
		},
	}

	// Check for outstanding debt to determine if we should create mandate with schedule
	schedule, err := s.generateScheduleData(ctx, clientID)
	if err != nil {
		return PendingCollection{}, err
	}

	// If there is outstanding debt, add schedules to mandate request
	var pc PendingCollection
	if schedule.Amount > 0 {
		mandateRequest.Schedules = []allpay.Schedule{{
			ScheduleDate:  schedule.CollectionDate.Format("2006-01-02"),
			Amount:        schedule.Amount,
			Frequency:     "1",
			TotalPayments: 1,
		}}

		// Create pending collection record
		var cd pgtype.Date
		_ = cd.Scan(schedule.CollectionDate)

		err = tx.CreatePendingCollection(ctx, store.CreatePendingCollectionParams{
			ClientID:       clientID,
			CollectionDate: cd,
			Amount:         schedule.Amount,
			CreatedBy:      ctx.(auth.Context).User.ID,
		})
		if err != nil {
			s.Logger(ctx).Error(fmt.Sprintf("Error creating pending collection for client : %d", clientID), slog.String("err", err.Error()))
			return PendingCollection{}, err
		}

		pc = schedule
	}

	err = s.allpay.CreateMandate(ctx, mandateRequest)

	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error creating mandate with allpay, rolling back payment method change for client : %d", clientID), slog.String("err", err.Error()))
		return PendingCollection{}, apierror.BadRequestError("Allpay", "Failed", err)
	}

	err = s.dispatch.PaymentMethodChanged(ctx, event.PaymentMethod{
		ClientID:      int(clientID),
		PaymentMethod: shared.PaymentMethodDirectDebit,
	})
	if err != nil {
		return PendingCollection{}, err
	}

	return pc, tx.Commit(ctx)
}
