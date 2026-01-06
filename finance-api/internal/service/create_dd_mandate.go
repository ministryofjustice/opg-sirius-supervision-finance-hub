package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
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

	err = s.dispatch.PaymentMethodChanged(ctx, event.PaymentMethod{
		ClientID:      int(clientID),
		PaymentMethod: shared.PaymentMethodDirectDebit,
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
