package service

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log/slog"
)

func (s *Service) GetPermittedAdjustments(ctx context.Context, invoiceId int32) ([]shared.AdjustmentType, error) {
	balance, err := s.store.GetInvoiceBalanceDetails(ctx, invoiceId)

	if err != nil {
		s.Logger(ctx).Error("Error get invoice balance in get permitted adjustments", slog.String("err", err.Error()))
		return nil, err
	}

	var permitted []shared.AdjustmentType
	if balance.WriteOffAmount == 0 {
		permitted = append(permitted, shared.AdjustmentTypeCreditMemo)
		if balance.Outstanding > 0 {
			permitted = append(permitted, shared.AdjustmentTypeWriteOff)
		}
		if (balance.Feetype == "AD" && balance.Outstanding < 10000) || (balance.Feetype != "AD" && balance.Outstanding < 32000) {
			permitted = append(permitted, shared.AdjustmentTypeDebitMemo)
		}
	} else {
		permitted = append(permitted, shared.AdjustmentTypeWriteOffReversal)
	}

	feeReductionDetails, err := s.store.GetInvoiceFeeReductionReversalDetails(ctx, invoiceId)
	if err != nil {
		s.Logger(ctx).Error("Error get invoice fee reduction details in get permitted adjustments", slog.String("err", err.Error()))
		return nil, err
	}

	// add condition for fee reduction reversal
	// if there is an existing fee reduction credit on the invoice and there is not an equivalent fee reduction reversal for that fee reductio
	// (calculated by the total of fee reduction credits + total of fee reduction reversal debits being more than zero)

	if feeReductionDetails.ReversalTotal < feeReductionDetails.FeeReductionTotal {
		permitted = append(permitted, shared.AdjustmentTypeFeeReductionReversal)
	}

	return permitted, nil
}
