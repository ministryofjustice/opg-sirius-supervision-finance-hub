//go:build seed && !release

package seed

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *seeder) fixture2(ctx context.Context) {
	c1 := client{
		courtRef: "abc",
		person: person{
			firstName: "John",
			surname:   "Smith",
		},
	}
	c2 := client{
		courtRef: "def",
		person: person{
			firstName: "Jane",
			surname:   "Smythe",
		},
	}
	d1 := deputy{
		deputyType: "LAY",
		client:     &c1,
		person: person{
			firstName: "Jim",
			surname:   "Smoth",
		},
	}
	s.publicClient.createClient(ctx, &c1)
	s.publicClient.createClient(ctx, &c2)
	s.publicClient.createDeputy(ctx, d1)

	i1 := shared.AddManualInvoice{
		InvoiceType:      shared.InvoiceTypeS2,
		Amount:           shared.Nillable[int]{Value: 32000, Valid: true},
		RaisedDate:       shared.Nillable[shared.Date]{Value: shared.NewDate("2024-01-01"), Valid: true},
		StartDate:        shared.Nillable[shared.Date]{Value: shared.NewDate("2024-01-01"), Valid: true},
		EndDate:          shared.Nillable[shared.Date]{Value: shared.NewDate("2024-01-01"), Valid: true},
		SupervisionLevel: shared.Nillable[string]{Value: "GENERAL", Valid: true},
	}

	_ = s.paymentsClient.AddManualInvoice(ctx, c1.id, i1)
}
