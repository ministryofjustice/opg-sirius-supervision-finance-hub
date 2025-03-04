package service

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) GetPermittedAdjustments(ctx context.Context, invoiceId int32) ([]shared.AdjustmentType, error) {
	balance, err := s.store.GetInvoiceBalanceDetails(ctx, invoiceId)
	if err != nil {
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

	return permitted, nil
}
