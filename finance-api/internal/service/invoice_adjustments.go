package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/shared"
	"math"
)

func (s *Service) GetInvoiceAdjustments(ctx context.Context, clientId int) (*shared.InvoiceAdjustments, error) {
	var adjustments shared.InvoiceAdjustments

	data, err := s.store.GetInvoiceAdjustments(ctx, int32(clientId))
	if err != nil {
		return nil, err
	}

	for _, ia := range data {
		var a = shared.InvoiceAdjustment{
			Id:             int(ia.ID),
			InvoiceRef:     ia.InvoiceRef,
			RaisedDate:     shared.Date{Time: ia.RaisedDate.Time},
			Amount:         int(math.Abs(float64(ia.Amount))),
			AdjustmentType: shared.ParseAdjustmentType(ia.AdjustmentType),
			Notes:          ia.Notes,
			Status:         ia.Status,
		}

		adjustments = append(adjustments, a)
	}

	return &adjustments, nil
}
