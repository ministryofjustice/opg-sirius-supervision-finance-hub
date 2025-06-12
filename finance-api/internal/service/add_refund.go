package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log/slog"
)

func (s *Service) AddRefund(ctx context.Context, clientId int32, refund shared.AddRefund) error {
	refundableAmount, err := s.store.GetRefundAmount(ctx, clientId)
	if err != nil {
		return err
	}

	if refundableAmount == 0 {
		s.Logger(ctx).Error("Error creating refund with no funds to refund")
		return apierror.BadRequest{Reason: "NoCreditToRefund"}
	}

	var (
		clientID      pgtype.Int4
		amount        pgtype.Int4
		notes         pgtype.Text
		accountName   pgtype.Text
		accountNumber pgtype.Text
		sortCode      pgtype.Text
		createdBy     pgtype.Int4
	)

	_ = store.ToInt4(&clientID, clientId)
	_ = store.ToInt4(&amount, refundableAmount)
	_ = notes.Scan(refund.RefundNotes)
	_ = accountName.Scan(refund.AccountName)
	_ = accountNumber.Scan(refund.AccountNumber)
	_ = sortCode.Scan(refund.SortCode)
	_ = store.ToInt4(&createdBy, ctx.(auth.Context).User.ID)

	_, err = s.store.CreateRefund(ctx, store.CreateRefundParams{
		ClientID:      clientID,
		Amount:        amount,
		Notes:         notes,
		CreatedBy:     createdBy,
		AccountName:   accountName,
		AccountNumber: accountNumber,
		SortCode:      sortCode,
	})
	if err != nil {
		s.Logger(ctx).Error("Error creating refund", slog.String("err", err.Error()))
		return err
	}

	return s.dispatch.RefundAdded(ctx, event.RefundAdded{ClientID: clientId})
}
