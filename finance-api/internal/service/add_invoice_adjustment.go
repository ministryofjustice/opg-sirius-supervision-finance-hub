package service

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log/slog"
)

func (s *Service) AddInvoiceAdjustment(ctx context.Context, clientId int32, invoiceId int32, adjustment *shared.AddInvoiceAdjustmentRequest) (*shared.InvoiceReference, error) {
	balance, err := s.store.GetInvoiceBalanceDetails(ctx, invoiceId)
	if err != nil {
		s.Logger(ctx).Error("Get invoice balance has an issue " + err.Error())
		return nil, err
	}

	clientInfo, err := s.store.GetAccountInformation(ctx, clientId)
	if err != nil {
		s.Logger(ctx).Error("Get account information has an issue " + err.Error())
		return nil, err
	}

	feeReductionDetails, err := s.store.GetInvoiceFeeReductionReversalDetails(ctx, invoiceId)

	err = s.validateAdjustmentAmount(ctx, adjustment, balance, feeReductionDetails)
	if err != nil {
		return nil, err
	}

	params := store.CreatePendingInvoiceAdjustmentParams{
		ClientID:       clientId,
		InvoiceID:      invoiceId,
		AdjustmentType: adjustment.AdjustmentType.Key(),
		Amount:         s.calculateAdjustmentAmount(adjustment, balance, clientInfo.Credit),
		Notes:          adjustment.AdjustmentNotes,
		CreatedBy:      ctx.(auth.Context).User.ID,
	}
	invoiceReference, err := s.store.CreatePendingInvoiceAdjustment(ctx, params)
	if err != nil {
		s.Logger(ctx).Error("Error creating pending invoice adjustment", slog.String("err", err.Error()))
		return nil, err
	}

	return &shared.InvoiceReference{InvoiceRef: invoiceReference}, nil
}

func (s *Service) validateAdjustmentAmount(ctx context.Context, adjustment *shared.AddInvoiceAdjustmentRequest, balance store.GetInvoiceBalanceDetailsRow, feeReductionDetails store.GetInvoiceFeeReductionReversalDetailsRow) error {
	switch adjustment.AdjustmentType {
	case shared.AdjustmentTypeCreditMemo:
		if adjustment.Amount-balance.Outstanding > balance.Initial {
			return apierror.BadRequestError("Amount", fmt.Sprintf("Amount entered must be equal to or less than £%s", shared.IntToDecimalString(int(balance.Initial+balance.Outstanding))), nil)
		}
	case shared.AdjustmentTypeDebitMemo:
		var maxBalance int32
		if balance.Feetype == "AD" {
			maxBalance = 10000
		} else {
			maxBalance = 32000
		}
		if adjustment.Amount+balance.Outstanding > maxBalance {
			return apierror.BadRequestError("Amount", fmt.Sprintf("Amount entered must be equal to or less than £%s", shared.IntToDecimalString(int(maxBalance-balance.Outstanding))), nil)
		}
	case shared.AdjustmentTypeWriteOff:
		if balance.Outstanding < 1 {
			return apierror.BadRequestError("Amount", "No outstanding balance to write off", nil)
		}
	case shared.AdjustmentTypeWriteOffReversal:
		if ctx.(auth.Context).User.IsFinanceManager() {
			if adjustment.ManagerOverride && adjustment.Amount == 0 {
				return apierror.BadRequestError("Amount", "The write-off reversal amount cannot be zero", nil)
			}
			if adjustment.Amount > balance.WriteOffAmount {
				return apierror.BadRequestError("Amount", fmt.Sprintf("The write-off reversal amount must be £%s or less", shared.IntToDecimalString(int(balance.WriteOffAmount))), nil)
			}
		} else {
			if adjustment.Amount > 0 {
				return apierror.BadRequestError("Amount", "Insufficient authorisation to manually adjust write off reversal amount", nil)
			}
			if balance.WriteOffAmount == 0 {
				return apierror.BadRequestError("Amount", "A write off reversal cannot be added to an invoice without an associated write off", nil)
			}
		}
	case shared.AdjustmentTypeFeeReductionReversal:
		unreversedFeeReductionTotal := feeReductionDetails.FeeReductionTotal.Int64 - feeReductionDetails.ReversalTotal.Int64
		if int64(adjustment.Amount) > unreversedFeeReductionTotal {
			return apierror.BadRequestError("Amount", fmt.Sprintf("The fee reduction reversal amount must be £%s or less", shared.IntToDecimalString(int(unreversedFeeReductionTotal))), nil)
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
		return -adjustment.Amount
	case shared.AdjustmentTypeWriteOffReversal:
		if adjustment.Amount > 0 {
			return -adjustment.Amount
		}
		if balance.WriteOffAmount > customerCreditBalance {
			return -customerCreditBalance
		}
		return -balance.WriteOffAmount
	default:
		return adjustment.Amount
	}
}
