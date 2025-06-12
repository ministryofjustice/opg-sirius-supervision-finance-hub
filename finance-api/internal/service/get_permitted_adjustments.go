package service

import (
	"context"
	"fmt"
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
	fmt.Println(feeReductionDetails)
	if err != nil {
		s.Logger(ctx).Error("Error get invoice fee reduction details in get permitted adjustments", slog.String("err", err.Error()))
		return nil, err
	}

	if feeReductionDetails.FeeReductionTotal.Int64 > feeReductionDetails.ReversalTotal.Int64 {
		permitted = append(permitted, shared.AdjustmentTypeFeeReductionReversal)
	}

	return permitted, nil
}
