package service

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log/slog"
)

func (s *Service) GetInvoiceAdjustments(ctx context.Context, clientId int32) (shared.InvoiceAdjustments, error) {
	var adjustments shared.InvoiceAdjustments

	data, err := s.store.GetInvoiceAdjustments(ctx, clientId)

	if err != nil {
		s.Logger(ctx).Error("Error get invoice adjustments", slog.String("err", err.Error()))
		return adjustments, err
	}

	for _, ia := range data {
		var a = shared.InvoiceAdjustment{
			Id:             int(ia.ID),
			InvoiceRef:     ia.InvoiceRef,
			RaisedDate:     shared.Date{Time: ia.RaisedDate.Time},
			Amount:         int(ia.Amount),
			AdjustmentType: shared.ParseAdjustmentType(ia.AdjustmentType),
			Notes:          ia.Notes,
			Status:         ia.Status,
			CreatedBy:      int(ia.CreatedBy),
		}

		adjustments = append(adjustments, a)
	}

	return adjustments, nil
}
