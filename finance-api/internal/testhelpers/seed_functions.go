package testhelpers

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"time"
)

func (s *Seeder) CreateClient(ctx context.Context, firstName string, surname string, courtRef string, sopNumber string) int32 {
	var clientId int32
	err := s.Conn.QueryRow(ctx, "INSERT INTO public.persons VALUES (NEXTVAL('public.persons_id_seq'), $1, $2, $3) RETURNING id", firstName, surname, courtRef).Scan(&clientId)
	assert.NoError(s.t, err, "failed to add Person: %v", err)
	_, err = s.Conn.Exec(ctx, "INSERT INTO supervision_finance.finance_client VALUES ($1, $1, $2, 'DEMANDED', NULL, $3) RETURNING id", clientId, sopNumber, courtRef)
	assert.NoError(s.t, err, "failed to add FinanceClient: %v", err)
	return clientId
}

func (s *Seeder) CreateDeputy(ctx context.Context, clientId int32, firstName string, surname string, deputyType string) int32 {
	var deputyId int32
	err := s.Conn.QueryRow(ctx, "INSERT INTO public.persons VALUES (NEXTVAL('public.persons_id_seq'), $1, $2, NULL, $3, $4) RETURNING id", firstName, surname, clientId, deputyType).Scan(&deputyId)
	assert.NoError(s.t, err, "failed to add Deputy: %v", err)
	_, err = s.Conn.Exec(ctx, "UPDATE public.persons SET feepayer_id = $1 WHERE id = $2", deputyId, clientId)
	assert.NoError(s.t, err, "failed to add Deputy to FinanceClient: %v", err)
	return deputyId
}

func (s *Seeder) CreateOrder(ctx context.Context, clientId int32, status string) {
	_, err := s.Conn.Exec(ctx, "INSERT INTO public.cases VALUES (NEXTVAL('public.cases_id_seq'), $1, $2)", clientId, status)
	assert.NoError(s.t, err, "failed to add order: %v", err)
}

func (s *Seeder) CreateTestAssignee(ctx context.Context) {
	_, err := s.Conn.Exec(ctx, "INSERT INTO public.assignees VALUES (NEXTVAL('public.assignees_id_seq'), $1, $2)", "Johnny", "Test")
	assert.NoError(s.t, err, "failed to create test assignee: %v", err)
}

func (s *Seeder) CreateInvoice(ctx context.Context, clientID int32, invoiceType shared.InvoiceType, amount *string, raisedDate *string, startDate *string, endDate *string, supervisionLevel *string, createdDate *string) (int32, string) {
	invoice := shared.AddManualInvoice{
		InvoiceType:      invoiceType,
		Amount:           shared.TransformNillableInt(amount),
		RaisedDate:       shared.TransformNillableDate(raisedDate),
		StartDate:        shared.TransformNillableDate(startDate),
		EndDate:          shared.TransformNillableDate(endDate),
		SupervisionLevel: shared.TransformNillableString(supervisionLevel),
	}

	err := s.Service.AddManualInvoice(ctx, clientID, invoice)
	assert.NoError(s.t, err, "failed to add invoice: %v", err)

	var id int32
	var reference string
	err = s.Conn.QueryRow(ctx, "SELECT id, reference FROM supervision_finance.invoice ORDER BY id DESC LIMIT 1").Scan(&id, &reference)
	assert.NoError(s.t, err, "failed find created invoice: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger SET datetime = $1, created_at = $1 WHERE id IN (SELECT ledger_id FROM supervision_finance.ledger_allocation WHERE invoice_id = $2)", raisedDate, id)
	assert.NoError(s.t, err, "failed to update ledger dates for reduction: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger_allocation SET datetime = $1 WHERE invoice_id = $2", raisedDate, id)
	assert.NoError(s.t, err, "failed to update ledger allocation dates for reduction: %v", err)

	if createdDate != nil {
		_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.invoice SET created_at = $2 WHERE id = $1", id, &createdDate)
		assert.NoError(s.t, err, "failed to update created date: %v", err)
	}

	return id, reference
}

func (s *Seeder) CreatePendingAdjustment(ctx context.Context, clientID int32, invoiceId int32, adjustmentType shared.AdjustmentType, amount int32, notes string) {
	adjustment := shared.AddInvoiceAdjustmentRequest{
		AdjustmentType:  adjustmentType,
		AdjustmentNotes: notes,
		Amount:          amount,
	}

	_, err := s.Service.AddInvoiceAdjustment(ctx, clientID, invoiceId, &adjustment)
	assert.NoError(s.t, err, "failed to add adjustment: %v", err)
}

func (s *Seeder) CreateAdjustment(ctx context.Context, clientID int32, invoiceId int32, adjustmentType shared.AdjustmentType, amount int32, notes string, approvedDate *time.Time) {
	adjustment := shared.AddInvoiceAdjustmentRequest{
		AdjustmentType:  adjustmentType,
		AdjustmentNotes: notes,
		Amount:          amount,
	}

	_, err := s.Service.AddInvoiceAdjustment(ctx, clientID, invoiceId, &adjustment)
	assert.NoError(s.t, err, "failed to add adjustment: %v", err)

	var id int32
	err = s.Conn.QueryRow(ctx, "SELECT id FROM supervision_finance.invoice_adjustment ORDER BY id DESC LIMIT 1").Scan(&id)
	assert.NoError(s.t, err, "failed find created adjustment: %v", err)

	var maxLedger int
	_ = s.Conn.QueryRow(ctx, "SELECT MAX(id) FROM supervision_finance.ledger").Scan(&maxLedger)

	err = s.Service.UpdatePendingInvoiceAdjustment(ctx, clientID, id, shared.AdjustmentStatusApproved)
	assert.NoError(s.t, err, "failed to approve adjustment: %v", err)

	if approvedDate != nil {
		_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger SET created_at = $1 WHERE id > $2", &approvedDate, maxLedger)
		assert.NoError(s.t, err, "failed to update created date: %v", err)
	}
}

func (s *Seeder) ApproveAdjustment(ctx context.Context, clientID int32, adjustmentId int32) {
	err := s.Service.UpdatePendingInvoiceAdjustment(ctx, clientID, adjustmentId, shared.AdjustmentStatusApproved)
	assert.NoError(s.t, err, "failed to approve adjustment: %v", err)
}

func (s *Seeder) CreateFeeReduction(ctx context.Context, clientId int32, feeType shared.FeeReductionType, startYear string, length int, notes string, createdAt time.Time) {
	received := shared.NewDate(startYear + "-01-01")
	reduction := shared.AddFeeReduction{
		FeeType:       feeType,
		StartYear:     startYear,
		LengthOfAward: length,
		DateReceived:  &received,
		Notes:         notes,
	}
	err := s.Service.AddFeeReduction(ctx, clientId, reduction)
	assert.NoError(s.t, err, "failed to create fee reduction: %v", err)

	// update created dates to enable testing reductions added in the past
	var id int
	err = s.Conn.QueryRow(ctx, "SELECT id FROM supervision_finance.fee_reduction ORDER BY id DESC LIMIT 1").Scan(&id)
	assert.NoError(s.t, err, "failed find created reduction: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger SET datetime = $1, created_at = $1 WHERE fee_reduction_id = $2", createdAt, id)
	assert.NoError(s.t, err, "failed to update ledger dates for reduction: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger_allocation SET datetime = $1 WHERE ledger_id IN (SELECT id FROM supervision_finance.ledger WHERE fee_reduction_id = $2)", createdAt, id)
	assert.NoError(s.t, err, "failed to update ledger allocation dates for reduction: %v", err)
}

func (s *Seeder) CreateChequePayment(ctx context.Context, amount int32, bankDate time.Time, courtRef string, pisNumber int, uploadDate time.Time) {
	payment := shared.PaymentDetails{
		Amount:       amount,
		BankDate:     bankDate,
		CourtRef:     courtRef,
		LedgerType:   shared.TransactionTypeSupervisionChequePayment.Key(),
		ReceivedDate: uploadDate,
		PisNumber:    shared.Nillable[int]{Value: pisNumber, Valid: true},
	}

	tx, err := s.Service.BeginStoreTx(ctx)
	assert.NoError(s.t, err, "failed to begin transaction: %v", err)

	var latestLedgerId int
	err = s.Conn.QueryRow(ctx, "SELECT COALESCE(MAX(id), 0) FROM supervision_finance.ledger").Scan(&latestLedgerId)
	assert.NoError(s.t, err, "failed to find latest ledger id: %v", err)

	failedLines := make(map[int]string)

	err = s.Service.ProcessPaymentsUploadLine(ctx, tx, payment, 0, &failedLines)
	assert.NoError(s.t, err, "payment not processed: %v", err)
	assert.Len(s.t, failedLines, 0, "payment failed: %v", failedLines)

	err = tx.Commit(ctx)
	assert.NoError(s.t, err, "failed to commit payment: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger SET datetime = $1, created_at = $1 WHERE id > $2", uploadDate, latestLedgerId)
	assert.NoError(s.t, err, "failed to update ledger dates for payment: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger_allocation SET datetime = $1 WHERE ledger_id > $2", uploadDate, latestLedgerId)
	assert.NoError(s.t, err, "failed to update ledger allocation dates for payment: %v", err)

	var newMaxLedger int
	err = s.Conn.QueryRow(ctx, "SELECT COALESCE(MAX(id), 0) FROM supervision_finance.ledger").Scan(&newMaxLedger)
	assert.NoError(s.t, err, "failed to find latest ledger id: %v", err)

	assert.Greater(s.t, newMaxLedger, latestLedgerId, "no ledgers created")
}

func (s *Seeder) CreatePayment(ctx context.Context, amount int32, bankDate time.Time, courtRef string, ledgerType shared.TransactionType, uploadDate time.Time) {
	payment := shared.PaymentDetails{
		Amount:       amount,
		BankDate:     bankDate,
		CourtRef:     courtRef,
		LedgerType:   ledgerType.Key(),
		ReceivedDate: uploadDate,
	}

	tx, err := s.Service.BeginStoreTx(ctx)
	assert.NoError(s.t, err, "failed to begin transaction: %v", err)

	var latestLedgerId int
	err = s.Conn.QueryRow(ctx, "SELECT COALESCE(MAX(id), 0) FROM supervision_finance.ledger").Scan(&latestLedgerId)
	assert.NoError(s.t, err, "failed to find latest ledger id: %v", err)

	failedLines := make(map[int]string)

	err = s.Service.ProcessPaymentsUploadLine(ctx, tx, payment, 0, &failedLines)
	assert.NoError(s.t, err, "payment not processed: %v", err)
	assert.Len(s.t, failedLines, 0, "payment failed: %v", failedLines)

	err = tx.Commit(ctx)
	assert.NoError(s.t, err, "failed to commit payment: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger SET datetime = $1, created_at = $1 WHERE id > $2", uploadDate, latestLedgerId)
	assert.NoError(s.t, err, "failed to update ledger dates for payment: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger_allocation SET datetime = $1 WHERE ledger_id > $2", uploadDate, latestLedgerId)
	assert.NoError(s.t, err, "failed to update ledger allocation dates for payment: %v", err)

	var newMaxLedger int
	err = s.Conn.QueryRow(ctx, "SELECT COALESCE(MAX(id), 0) FROM supervision_finance.ledger").Scan(&newMaxLedger)
	assert.NoError(s.t, err, "failed to find latest ledger id: %v", err)

	assert.Greater(s.t, newMaxLedger, latestLedgerId, "no ledgers created")
}

type FeeRange struct {
	FromDate         time.Time
	ToDate           time.Time
	SupervisionLevel string
	Amount           int
}

func (s *Seeder) AddFeeRanges(ctx context.Context, invoiceId int32, ranges []FeeRange) {
	for _, r := range ranges {

		_, err := s.Conn.Exec(ctx,
			"INSERT INTO supervision_finance.invoice_fee_range VALUES (NEXTVAL('supervision_finance.invoice_fee_range_id_seq'), $1, $2, $3, $4, $5)",
			invoiceId, r.SupervisionLevel, r.FromDate, r.ToDate, r.Amount)
		assert.NoError(s.t, err, "failed to add fee range: %v", err)
	}
}
