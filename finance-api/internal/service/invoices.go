package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/shared"
)

func (s *Service) GetInvoices(id int) (*shared.Invoices, error) {
	ctx := context.Background()

	fc, err := s.Store.GetInvoices(ctx, pgtype.Int4{Int32: int32(id), Valid: true})
	if err != nil {
		return nil, err
	}

	var invoices shared.Invoices

	for _, i := range fc {
		var invoice = shared.Invoice{
			Id:                 int(i.ID),
			Ref:                i.Reference,
			Status:             "UNKNOWN",
			Amount:             int(i.Amount),
			RaisedDate:         shared.Date{Time: i.Raiseddate.Time},
			Received:           0,
			OutstandingBalance: int(i.Cacheddebtamount.Int32),
			Ledgers:            nil,
			SupervisionLevels:  nil,
		}
		invoices = append(invoices, invoice)
	}

	return &invoices, nil
}
