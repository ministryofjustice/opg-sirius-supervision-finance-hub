package testhelpers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (s *Seeder) CreateClient(ctx context.Context, firstName string, surname string, courtRef string, sopNumber string, status string) int32 {
	var clientId int32
	err := s.Conn.QueryRow(ctx, "INSERT INTO public.persons (id, firstname, surname, caserecnumber, type, clientstatus) VALUES (NEXTVAL('public.persons_id_seq'), $1, $2, $3, 'actor_client', $4) RETURNING id", firstName, surname, courtRef, status).Scan(&clientId)
	assert.NoError(s.t, err, "failed to add Person: %v", err)
	_, err = s.Conn.Exec(ctx, "INSERT INTO supervision_finance.finance_client VALUES ($1, $1, $2, 'DEMANDED', NULL, $3) RETURNING id", clientId, sopNumber, courtRef)
	assert.NoError(s.t, err, "failed to add FinanceClient: %v", err)
	return clientId
}

func (s *Seeder) CreateAddresses(ctx context.Context, clientId int32, addressLines []string, town string, county string, postcode string, airmailRequired bool) int32 {
	var addressId int32
	addressLinesJson, _ := json.Marshal(addressLines)
	err := s.Conn.QueryRow(ctx, "INSERT INTO public.addresses VALUES (NEXTVAL('public.addresses_id_seq'), $1, $2, $3, $4, $5, $6) RETURNING id", clientId, addressLinesJson, town, county, postcode, airmailRequired).Scan(&addressId)
	assert.NoError(s.t, err, "failed to add Address: %v", err)
	return addressId
}

func (s *Seeder) CreateDeputy(ctx context.Context, clientId int32, firstName string, surname string, deputyType string) int32 {
	var deputyId int32
	err := s.Conn.QueryRow(ctx, "INSERT INTO public.persons (id, firstname, surname, feepayer_id, deputytype) VALUES (NEXTVAL('public.persons_id_seq'), $1, $2, $3, $4) RETURNING id", firstName, surname, clientId, deputyType).Scan(&deputyId)
	assert.NoError(s.t, err, "failed to add Deputy: %v", err)
	_, err = s.Conn.Exec(ctx, "UPDATE public.persons SET feepayer_id = $1 WHERE id = $2", deputyId, clientId)
	assert.NoError(s.t, err, "failed to add Deputy to FinanceClient: %v", err)
	return deputyId
}

func (s *Seeder) CreateOrder(ctx context.Context, clientId int32) {
	_, err := s.Conn.Exec(ctx, "INSERT INTO public.cases (id, client_id, orderstatus) VALUES (NEXTVAL('public.cases_id_seq'), $1, 'ACTIVE')", clientId)
	assert.NoError(s.t, err, "failed to add order: %v", err)
}

func (s *Seeder) CreateClosedOrder(ctx context.Context, clientId int32, closedOn time.Time, reason string) {
	_, err := s.Conn.Exec(ctx, "INSERT INTO public.cases VALUES (NEXTVAL('public.cases_id_seq'), $1, 'CLOSED', $2, $3)", clientId, closedOn, reason)
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

	if createdDate == nil {
		createdDate = s.Today().StringPtr()
	}

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger SET datetime = $1, created_at = $2 WHERE id IN (SELECT ledger_id FROM supervision_finance.ledger_allocation WHERE invoice_id = $3)", raisedDate, createdDate, id)
	assert.NoError(s.t, err, "failed to update ledger dates for reduction: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger_allocation SET datetime = $1 WHERE invoice_id = $2", raisedDate, id)
	assert.NoError(s.t, err, "failed to update ledger allocation dates for reduction: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.invoice SET created_at = $2 WHERE id = $1", id, &createdDate)
	assert.NoError(s.t, err, "failed to update created date: %v", err)

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

	var currentLedgerId int
	_ = s.Conn.QueryRow(ctx, "SELECT MAX(id) FROM supervision_finance.ledger").Scan(&currentLedgerId)

	_, err := s.Service.AddInvoiceAdjustment(ctx, clientID, invoiceId, &adjustment)
	assert.NoError(s.t, err, "failed to add adjustment: %v", err)

	var id int32
	err = s.Conn.QueryRow(ctx, "SELECT id FROM supervision_finance.invoice_adjustment ORDER BY id DESC LIMIT 1").Scan(&id)
	assert.NoError(s.t, err, "failed find created adjustment: %v", err)

	err = s.Service.UpdatePendingInvoiceAdjustment(ctx, clientID, id, shared.AdjustmentStatusApproved)
	assert.NoError(s.t, err, "failed to approve adjustment: %v", err)

	if approvedDate == nil {
		approvedDate = s.Today().DatePtr()
	}

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.invoice_adjustment SET created_at = $1 WHERE id = $2", approvedDate, id)
	assert.NoError(s.t, err, "failed to update approval dates: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger SET datetime = $1, created_at = $2 WHERE id > $3", approvedDate, approvedDate, currentLedgerId)
	assert.NoError(s.t, err, "failed to update ledger dates: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger_allocation SET datetime = $1 WHERE ledger_id > $2", approvedDate, currentLedgerId)
	assert.NoError(s.t, err, "failed to update ledger allocation dates: %v", err)
}

func (s *Seeder) CreateFeeReduction(ctx context.Context, clientId int32, feeType shared.FeeReductionType, startYear string, length int, notes string, createdAt time.Time) int32 {
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
	var id int32
	err = s.Conn.QueryRow(ctx, "SELECT id FROM supervision_finance.fee_reduction ORDER BY id DESC LIMIT 1").Scan(&id)
	assert.NoError(s.t, err, "failed find created reduction: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger SET datetime = $1, created_at = $1 WHERE fee_reduction_id = $2", createdAt, id)
	assert.NoError(s.t, err, "failed to update ledger dates for reduction: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger_allocation SET datetime = $1 WHERE ledger_id IN (SELECT id FROM supervision_finance.ledger WHERE fee_reduction_id = $2)", createdAt, id)
	assert.NoError(s.t, err, "failed to update ledger allocation dates for reduction: %v", err)

	return id
}

func (s *Seeder) CancelFeeReduction(ctx context.Context, feeReductionId int32) {
	err := s.Service.CancelFeeReduction(ctx, feeReductionId, shared.CancelFeeReduction{CancellationReason: ""})
	assert.NoError(s.t, err, "failed to cancel fee reduction: %v", err)
}

func (s *Seeder) CreatePayment(ctx context.Context, amount int32, bankDate time.Time, courtRef string, ledgerType shared.TransactionType, uploadDate time.Time, pisNumber int32) {
	payment := shared.PaymentDetails{
		Amount: amount,
		BankDate: pgtype.Date{
			Time:  bankDate,
			Valid: true,
		},
		CourtRef: pgtype.Text{
			String: courtRef,
			Valid:  true,
		},
		LedgerType: ledgerType,
		ReceivedDate: pgtype.Timestamp{
			Time:  uploadDate,
			Valid: true,
		},
		CreatedBy: pgtype.Int4{
			Int32: ctx.(auth.Context).User.ID,
			Valid: true,
		},
		PisNumber: pgtype.Int4{
			Int32: pisNumber,
			Valid: pisNumber > 0,
		},
	}

	tx, err := s.Service.BeginStoreTx(ctx)
	assert.NoError(s.t, err, "failed to begin transaction: %v", err)

	var latestLedgerId int
	err = s.Conn.QueryRow(ctx, "SELECT COALESCE(MAX(id), 0) FROM supervision_finance.ledger").Scan(&latestLedgerId)
	assert.NoError(s.t, err, "failed to find latest ledger id: %v", err)

	_, err = s.Service.ProcessPaymentsUploadLine(ctx, tx, payment)
	assert.NoError(s.t, err, "payment not processed: %v", err)

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

func (s *Seeder) ReversePayment(ctx context.Context, erroredCourtRef string, correctCourtRef string, amount string, bankDate time.Time, receivedDate time.Time, ledgerType shared.TransactionType, uploadDate time.Time) {
	var latestLedgerId int
	err := s.Conn.QueryRow(ctx, "SELECT COALESCE(MAX(id), 0) FROM supervision_finance.ledger").Scan(&latestLedgerId)
	assert.NoError(s.t, err, "failed to find latest ledger id: %v", err)

	records := [][]string{
		{"Payment type", "Current (errored) court reference", "New (correct) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
		{ledgerType.Key(), erroredCourtRef, correctCourtRef, bankDate.Format("02/01/2006"), receivedDate.Format("02/01/2006"), amount, ""},
	}
	failedLines, err := s.Service.ProcessPaymentReversals(ctx, records, shared.ReportTypeUploadMisappliedPayments, shared.Date{Time: uploadDate})
	assert.Empty(s.t, failedLines, "failed to process reversals")
	assert.NoError(s.t, err, "failed to process reversals: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger SET created_at = $1 WHERE id > $2", uploadDate, latestLedgerId)
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

func (s *Seeder) CreateWarning(ctx context.Context, personId int32, warningType string) {
	var warningId int
	err := s.Conn.QueryRow(ctx, "INSERT INTO public.warnings VALUES (NEXTVAL('public.warnings_id_seq'), $1, TRUE) RETURNING id", warningType).Scan(&warningId)
	assert.NoError(s.t, err, "failed to add warning: %v", err)

	_, err = s.Conn.Exec(ctx,
		"INSERT INTO public.person_warning VALUES ($1, $2)",
		personId, warningId)
	assert.NoError(s.t, err, "failed to add person warning: %v", err)
}

func (s *Seeder) CreateRefund(ctx context.Context, clientId int32, accountName string, accountNumber string, sortCode string, createdDate time.Time) int32 {
	err := s.Service.AddRefund(ctx, clientId, shared.AddRefund{
		AccountName:   accountName,
		AccountNumber: accountNumber,
		SortCode:      sortCode,
		RefundNotes:   "",
	})
	assert.NoError(s.t, err, "failed to add refund: %v", err)

	var id int32
	err = s.Conn.QueryRow(ctx, "SELECT id FROM supervision_finance.refund ORDER BY id DESC LIMIT 1").Scan(&id)
	assert.NoError(s.t, err, "failed find created refund: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.refund SET created_at = $1 WHERE id = $2", createdDate, id)
	assert.NoError(s.t, err, "failed to update refund date: %v", err)

	return id
}

func (s *Seeder) SetRefundDecision(ctx context.Context, clientId int32, refundId int32, decision shared.RefundStatus, decisionDate time.Time) {
	err := s.Service.UpdateRefundDecision(ctx, clientId, refundId, decision)
	assert.NoError(s.t, err, "failed to update refund decision: %v", err)

	if decision == shared.RefundStatusCancelled {
		_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.refund SET cancelled_at = $1 WHERE id = $2", decisionDate, refundId)
		assert.NoError(s.t, err, "failed to update refund cancelled date: %v", err)
	} else {
		_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.refund SET decision_at = $1 WHERE id = $2", decisionDate, refundId)
		assert.NoError(s.t, err, "failed to update refund date: %v", err)
	}
}

func (s *Seeder) ProcessApprovedRefunds(ctx context.Context, ids []int32, processingDate time.Time) {
	debtType := shared.DebtTypeApprovedRefunds
	report := shared.ReportRequest{
		ReportType: shared.ReportsTypeDebt,
		DebtType:   &debtType,
	}
	s.Service.PostReportActions(ctx, report)

	_, err := s.Conn.Exec(ctx, "UPDATE supervision_finance.refund SET processed_at = $1 WHERE id = ANY($2)", processingDate, ids)
	assert.NoError(s.t, err, "failed to update refund date: %v", err)
}

func (s *Seeder) FulfillRefund(ctx context.Context, refundId int32, amount int32, bankDate time.Time, courtRef string, accountName string, accountNumber string, sortCode string, uploadDate time.Time) {
	refund := shared.FulfilledRefundDetails{
		CourtRef: pgtype.Text{
			String: courtRef,
			Valid:  true,
		},
		Amount: pgtype.Int4{
			Int32: amount,
			Valid: true,
		},
		AccountName: pgtype.Text{
			String: accountName,
			Valid:  true,
		},
		AccountNumber: pgtype.Text{
			String: accountNumber,
			Valid:  true,
		},
		SortCode: pgtype.Text{
			String: sortCode,
			Valid:  true,
		},
		UploadedBy: pgtype.Int4{
			Int32: 10,
			Valid: true,
		},
		BankDate: pgtype.Date{
			Time:  bankDate,
			Valid: true,
		},
	}

	tx, err := s.Service.BeginStoreTx(ctx)
	assert.NoError(s.t, err, "failed to begin transaction: %v", err)

	var latestLedgerId int
	err = s.Conn.QueryRow(ctx, "SELECT COALESCE(MAX(id), 0) FROM supervision_finance.ledger").Scan(&latestLedgerId)
	assert.NoError(s.t, err, "failed to find latest ledger id: %v", err)

	_ = s.Service.ProcessFulfilledRefundsLine(ctx, tx, refundId, refund)
	assert.NoError(s.t, err, "refund not processed: %v", err)

	err = tx.Commit(ctx)
	assert.NoError(s.t, err, "failed to commit refund: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger SET datetime = $1, created_at = $1 WHERE id > $2", uploadDate, latestLedgerId)
	assert.NoError(s.t, err, "failed to update ledger dates for refund: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.ledger_allocation SET datetime = $1 WHERE ledger_id > $2", uploadDate, latestLedgerId)
	assert.NoError(s.t, err, "failed to update ledger allocation dates for refund: %v", err)

	_, err = s.Conn.Exec(ctx, "UPDATE supervision_finance.refund SET fulfilled_at = $1 WHERE id = $2", uploadDate, refundId)
	assert.NoError(s.t, err, "failed to update refund date: %v", err)

	var newMaxLedger int
	err = s.Conn.QueryRow(ctx, "SELECT COALESCE(MAX(id), 0) FROM supervision_finance.ledger").Scan(&newMaxLedger)
	assert.NoError(s.t, err, "failed to find latest ledger id: %v", err)

	assert.Greater(s.t, newMaxLedger, latestLedgerId, "no ledgers created")
}
