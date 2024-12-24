package testhelpers

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log"
)

func (s *Seeder) CreateClient(ctx context.Context, firstName string, surname string, courtRef string, sopNumber string) int {
	var (
		clientId int
	)
	err := s.Conn.QueryRow(ctx, "INSERT INTO public.persons VALUES (NEXTVAL('public.persons_id_seq'), $1, $2, $3) RETURNING id", firstName, surname, courtRef).Scan(&clientId)
	if err != nil {
		log.Fatalf("failed to add FinanceClient: %v", err)
	}
	_, err = s.Conn.Exec(ctx, "INSERT INTO supervision_finance.finance_client VALUES (NEXTVAL('supervision_finance.finance_client_id_seq'), $1, $2, 'DEMANDED') RETURNING id", clientId, sopNumber)
	if err != nil {
		log.Fatalf("failed to add finance_client: %v", err)
	}
	return clientId
}

func (s *Seeder) CreateDeputy(ctx context.Context, clientId int, firstName string, surname string, deputyType string) int {
	var (
		deputyId int
	)
	err := s.Conn.QueryRow(ctx, "INSERT INTO public.persons VALUES (NEXTVAL('public.persons_id_seq'), $1, $2, NULL, $3, $4) RETURNING id", firstName, surname, clientId, deputyType).Scan(&deputyId)
	if err != nil {
		log.Fatalf("failed to add Deputy: %v", err)
	}
	_, err = s.Conn.Exec(ctx, "UPDATE public.persons SET feepayer_id = $1 WHERE id = $2", deputyId, clientId)
	if err != nil {
		log.Fatalf("failed to add Deputy to FinanceClient: %v", err)
	}
	return deputyId
}

func (s *Seeder) CreateOrder(ctx context.Context, clientId int, status string) {
	_, err := s.Conn.Exec(ctx, "INSERT INTO public.cases VALUES (NEXTVAL('public.cases_id_seq'), $1, $2)", clientId, status)
	if err != nil {
		log.Fatalf("failed to add order: %v", err)
	}
}

func (s *Seeder) CreateInvoice(ctx context.Context, clientID int, invoiceType shared.InvoiceType, amount *string, raisedDate *string, startDate *string, endDate *string, supervisionLevel *string) (int, string) {
	invoice := shared.AddManualInvoice{
		InvoiceType:      invoiceType,
		Amount:           shared.TransformNillableInt(amount),
		RaisedDate:       shared.TransformNillableDate(raisedDate),
		StartDate:        shared.TransformNillableDate(startDate),
		EndDate:          shared.TransformNillableDate(endDate),
		SupervisionLevel: shared.TransformNillableString(supervisionLevel),
	}

	err := s.Service.AddManualInvoice(ctx, clientID, invoice)
	if err != nil {
		log.Fatalf("failed to add invoice: %v", err)
	}

	var (
		id        int
		reference string
	)
	err = s.Conn.QueryRow(ctx, "SELECT id, reference FROM supervision_finance.invoice ORDER BY id DESC LIMIT 1").Scan(&id, &reference)
	if err != nil {
		log.Fatalf("failed find created invoice: %v", err)
	}
	return id, reference
}

func (s *Seeder) CreateAdjustment(ctx context.Context, clientID int, invoiceId int, adjustmentType shared.AdjustmentType, amount int, notes string) int {
	adjustment := shared.AddInvoiceAdjustmentRequest{
		AdjustmentType:  adjustmentType,
		AdjustmentNotes: notes,
		Amount:          amount,
	}

	_, err := s.Service.AddInvoiceAdjustment(ctx, clientID, invoiceId, &adjustment)
	if err != nil {
		log.Fatalf("failed to add adjustment: %v", err)
	}

	var id int
	err = s.Conn.QueryRow(ctx, "SELECT id FROM supervision_finance.invoice_adjustment ORDER BY id DESC LIMIT 1").Scan(&id)
	if err != nil {
		log.Fatalf("failed find created adjustment: %v", err)
	}
	return id
}

func (s *Seeder) ApproveAdjustment(ctx context.Context, clientID int, adjustmentId int) {
	err := s.Service.UpdatePendingInvoiceAdjustment(ctx, clientID, adjustmentId, shared.AdjustmentStatusApproved)
	if err != nil {
		log.Fatalf("failed to approve adjustment: %v", err)
	}
}

func (s *Seeder) CreateFeeReduction(ctx context.Context, clientId int, feeType shared.FeeReductionType, startYear string, length int, notes string) {
	received := shared.NewDate(startYear + "-01-01")
	reduction := shared.AddFeeReduction{
		FeeType:       feeType,
		StartYear:     startYear,
		LengthOfAward: length,
		DateReceived:  &received,
		Notes:         notes,
	}
	err := s.Service.AddFeeReduction(ctx, clientId, reduction)
	if err != nil {
		log.Fatalf("failed to create fee reduction: %v", err)
	}
}
