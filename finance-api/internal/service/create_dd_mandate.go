package service

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"log/slog"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) CreateDirectDebitMandate(ctx context.Context, clientID int32, createMandate shared.CreateMandate) error {
	bankDetails := createMandate.BankAccount.BankDetails
	err := s.allpay.ModulusCheck(ctx, bankDetails.SortCode, bankDetails.AccountNumber)
	if err != nil {
		return apierror.BadRequestError("ModulusCheck", "Failed", err)
	}

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// update payment method first, in case this fails
	err = tx.UpdatePaymentMethod(ctx, store.UpdatePaymentMethodParams{
		PaymentMethod: shared.PaymentMethodDirectDebit.Key(),
		ClientID:      clientID,
	})
	if err != nil {
		return err
	}

	err = s.allpay.CreateMandate(ctx, &allpay.CreateMandateRequest{
		ClientReference: createMandate.ClientReference,
		Surname:         createMandate.Surname,
		Address: allpay.Address{
			Line1:    createMandate.Address.Line1,
			Town:     createMandate.Address.Town,
			PostCode: createMandate.Address.PostCode,
		},
		BankAccount: struct {
			BankDetails allpay.BankDetails `json:"BankDetails"`
		}{
			BankDetails: allpay.BankDetails{
				AccountName:   bankDetails.AccountName,
				SortCode:      bankDetails.SortCode,
				AccountNumber: bankDetails.AccountNumber,
			},
		},
	})

	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error creating mandate with allpay, rolling back payment method change for client : %d", clientID), slog.String("err", err.Error()))
		return apierror.BadRequestError("Allpay", "Failed", err)
	}

	//db entry to say we've recorded a new payment method
	_, err = tx.AddPaymentMethod(ctx, store.AddPaymentMethodParams{
		ClientID:  clientID,
		Type:      shared.PaymentMethodDirectDebit.Key(),
		CreatedBy: ctx.(auth.Context).User.ID,
	})

	if err != nil {
		s.Logger(ctx).Error("Updating payment method table had an issue " + err.Error())
		return err
	}

	err = s.dispatch.PaymentMethodChanged(ctx, event.PaymentMethod{
		ClientID:      int(clientID),
		PaymentMethod: shared.PaymentMethodDirectDebit,
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
