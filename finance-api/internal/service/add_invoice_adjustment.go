package service

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log/slog"
)

func (s *Service) AddInvoiceAdjustment(ctx context.Context, clientId int, invoiceId int, ledgerEntry *shared.AddInvoiceAdjustmentRequest) (*shared.InvoiceReference, error) {
	logger := telemetry.LoggerFromContext(ctx)

	balance, err := s.store.GetInvoiceBalanceDetails(ctx, int32(invoiceId))
	if err != nil {
		return nil, err
	}

	clientInfo, err := s.store.GetAccountInformation(ctx, int32(clientId))
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	err = s.validateAdjustmentAmount(ledgerEntry, balance)
	if err != nil {
		return nil, err
	}

	params := store.CreatePendingInvoiceAdjustmentParams{
		ClientID:       int32(clientId),
		InvoiceID:      int32(invoiceId),
		AdjustmentType: ledgerEntry.AdjustmentType.Key(),
		Amount:         s.calculateAdjustmentAmount(ledgerEntry, balance, clientInfo.Credit),
		Notes:          ledgerEntry.AdjustmentNotes,
		CreatedBy:      int32(ctx.(auth.Context).User.ID),
	}
	invoiceReference, err := s.store.CreatePendingInvoiceAdjustment(ctx, params)
	if err != nil {
		logger.Error("Error creating pending invoice adjustment", slog.String("err", err.Error()))
		return nil, err
	}

	return &shared.InvoiceReference{InvoiceRef: invoiceReference}, nil
}

func (s *Service) validateAdjustmentAmount(adjustment *shared.AddInvoiceAdjustmentRequest, balance store.GetInvoiceBalanceDetailsRow) error {
	switch adjustment.AdjustmentType {
	case shared.AdjustmentTypeCreditMemo:
		if int32(adjustment.Amount)-balance.Outstanding > balance.Initial {
			return apierror.BadRequestError("Amount", fmt.Sprintf("Amount entered must be equal to or less than £%s", shared.IntToDecimalString(int(balance.Initial+balance.Outstanding))), nil)
		}
	case shared.AdjustmentTypeDebitMemo:
		var maxBalance int32
		if balance.Feetype == "AD" {
			maxBalance = 10000
		} else {
			maxBalance = 32000
		}
		if int32(adjustment.Amount)+balance.Outstanding > maxBalance {
			return apierror.BadRequestError("Amount", fmt.Sprintf("Amount entered must be equal to or less than £%s", shared.IntToDecimalString(int(maxBalance-balance.Outstanding))), nil)
		}
	case shared.AdjustmentTypeWriteOff:
		if balance.Outstanding < 1 {
			return apierror.BadRequestError("Amount", "No outstanding balance to write off", nil)
		}
	case shared.AdjustmentTypeWriteOffReversal:
		if balance.WriteOffAmount == 0 {
			return apierror.BadRequest{Field: "Amount", Reason: "A write off reversal cannot be added to an invoice without an associated write off"}
		}
	default:
		return apierror.BadRequestError("AdjustmentType", "Unimplemented adjustment type", nil)
	}
	return nil
}

func (s *Service) calculateAdjustmentAmount(adjustment *shared.AddInvoiceAdjustmentRequest, balance store.GetInvoiceBalanceDetailsRow, customerCreditBalance int32) int32 {
	switch adjustment.AdjustmentType {
	case shared.AdjustmentTypeWriteOff:
		return balance.Outstanding
	case shared.AdjustmentTypeDebitMemo:
		return -int32(adjustment.Amount)
	case shared.AdjustmentTypeWriteOffReversal:
		if balance.WriteOffAmount > customerCreditBalance {
			return -customerCreditBalance
		}
		return -balance.WriteOffAmount
	default:
		return int32(adjustment.Amount)
	}
}
