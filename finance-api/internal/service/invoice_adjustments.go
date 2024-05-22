package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/shared"
)

func (s *Service) GetInvoiceAdjustments(clientID int) (*shared.InvoiceAdjustments, error) {
	ctx := context.Background()

	var adjustments shared.InvoiceAdjustments

	data, err := s.store.GetInvoiceAdjustments(ctx, int32(clientID))
	if err != nil {
		return nil, err
	}

	for _, ia := range data {
		var a = shared.InvoiceAdjustment{
			Id:             int(ia.ID),
			InvoiceRef:     ia.InvoiceRef,
			RaisedDate:     shared.Date{Time: ia.RaisedDate.Time},
			Amount:         int(ia.Amount),
			AdjustmentType: shared.ParseAdjustmentType(ia.Type),
			Notes:          ia.Notes.String,
			Status:         ia.Status,
		}

		adjustments = append(adjustments, a)
	}

	return &adjustments, nil
}
