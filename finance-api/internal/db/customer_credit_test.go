package db

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/testhelpers"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_customer_credit() {
	ctx := suite.ctx

	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)
	twoMonthsAgo := today.Sub(0, 2, 0)
	twoYearsAgo := today.Sub(2, 0, 0)
	threeYearsAgo := today.Sub(3, 0, 0)
	minimal := "10"

	// client 1 with:
	// - Credit balance due to overpayment
	// £100 - £223.45 = -£123.45
	client1ID := suite.seeder.CreateClient(ctx, "Ian", "Test", "12345678", "1234", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client1ID, "pfa")
	suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.seeder.CreatePayment(ctx, 22345, twoYearsAgo.Date(), "12345678", shared.TransactionTypeOPGBACSPayment, yesterday.Date(), 0)

	// client 2 with:
	// - Credit balance due to fee reduction
	// - Partially reapplied
	// £100 - £100 + £100 - £10 = -£90
	client2ID := suite.seeder.CreateClient(ctx, "John", "Suite", "87654321", "4321", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client2ID, "pfa")
	suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, twoYearsAgo.StringPtr(), nil, nil, nil, twoYearsAgo.StringPtr())
	suite.seeder.CreatePayment(ctx, 10000, twoYearsAgo.Date(), "87654321", shared.TransactionTypeOPGBACSPayment, twoYearsAgo.Date(), 0)
	_ = suite.seeder.CreateFeeReduction(ctx, client2ID, shared.FeeReductionTypeExemption, strconv.Itoa(threeYearsAgo.Date().Year()), 2, "A reduction", threeYearsAgo.Date())
	suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeS3, &minimal, yesterday.StringPtr(), yesterday.StringPtr(), nil, nil, yesterday.StringPtr())

	// Doesn't display client with:
	// - No credit balance after unapplied funds fully reapplied
	// £100 - £150 + £100 = £50 (outstanding)
	client3ID := suite.seeder.CreateClient(ctx, "Billy", "Client", "23456789", "2345", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client3ID, "pfa")
	suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, twoMonthsAgo.StringPtr(), nil, nil, nil, twoMonthsAgo.StringPtr())
	suite.seeder.CreatePayment(ctx, 15000, yesterday.Date(), "23456789", shared.TransactionTypeOPGBACSPayment, yesterday.Date(), 0)
	suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())

	// Only show debt up to 'to' date
	// £100 - £200 = -£100
	// CCB applied to invoice but after 'to' date
	client4ID := suite.seeder.CreateClient(ctx, "Polly", "Partial", "34567890", "3456", "ACTIVE")
	suite.seeder.CreateOrder(ctx, client4ID, "pfa")
	suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeAD, nil, twoYearsAgo.StringPtr(), nil, nil, nil, twoYearsAgo.StringPtr())
	suite.seeder.CreatePayment(ctx, 20000, twoYearsAgo.Date(), "34567890", shared.TransactionTypeOPGBACSPayment, twoYearsAgo.Date(), 0)
	suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeS3, &minimal, today.StringPtr(), today.StringPtr(), nil, nil, today.StringPtr())

	c := Client{suite.seeder.Conn}

	to := shared.NewDate(yesterday.String())

	rows, err := c.Run(ctx, NewCustomerCredit(CustomerCreditInput{
		ToDate: &to,
	}))

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 4, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// client 1
	assert.Equal(suite.T(), "Ian Test", results[0]["Customer name"], "Customer name - client 1")
	assert.Equal(suite.T(), "12345678", results[0]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "1234", results[0]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), "123.45", results[0]["Credit balance"], "Credit balance - client 1")

	// client 2
	assert.Equal(suite.T(), "Polly Partial", results[1]["Customer name"], "Customer name - client 2")
	assert.Equal(suite.T(), "34567890", results[1]["Customer number"], "Customer number - client 2")
	assert.Equal(suite.T(), "3456", results[1]["SOP number"], "SOP number - client 2")
	assert.Equal(suite.T(), "100.00", results[1]["Credit balance"], "Credit balance - client 2")

	// client 3
	assert.Equal(suite.T(), "John Suite", results[2]["Customer name"], "Customer name - client 3")
	assert.Equal(suite.T(), "87654321", results[2]["Customer number"], "Customer number - client 3")
	assert.Equal(suite.T(), "4321", results[2]["SOP number"], "SOP number - client 3")
	assert.Equal(suite.T(), "90.00", results[2]["Credit balance"], "Credit balance - client 3")
}

func (suite *IntegrationSuite) Test_customer_credit_uses_correct_ledger_date() {
	ctx := suite.ctx
	today := suite.seeder.Today()
	yesterday := today.Sub(0, 0, 1)

	// ledger will compare using the ledger created_at where it exists
	// this ledger amount will be included in the credit amount on the report as the ledger created_at (yesterday) is BEFORE the report run date
	clientID := suite.seeder.CreateClient(ctx, "Ianthe", "0", "12345679", "1233", "ACTIVE")
	suite.seeder.CreateInvoice(ctx, clientID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.CreatePaymentForCustomerCreditTest(ctx, "12345679", "UNAPPLIED", today, true, yesterday)

	// ledger will compare using the ledger created_at where it exists
	// this ledger amount will NOT be included in the credit amount on the report as the ledger created_at (today) is AFTER the report run date
	client1ID := suite.seeder.CreateClient(ctx, "Bernard", "1", "12345678", "1234", "ACTIVE")
	suite.seeder.CreateInvoice(ctx, client1ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.CreatePaymentForCustomerCreditTest(ctx, "12345678", "UNAPPLIED", yesterday, true, today)

	// ledger will compare using the ledger datetime when created_at is null
	// this ledger amount will be included in the credit amount on the report as the ledger created_at is null and the ledger date_time (yesterday) is BEFORE the report run date
	client2ID := suite.seeder.CreateClient(ctx, "Beryl", "2", "22345678", "2234", "ACTIVE")
	suite.seeder.CreateInvoice(ctx, client2ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.CreatePaymentForCustomerCreditTest(ctx, "22345678", "UNAPPLIED", yesterday, false, testhelpers.DateHelper{})

	// ledger will compare using the ledger datetime when created_at is null
	// this ledger amount will NOT be included in the credit amount on the report as the ledger created_at is null and the ledger date_time (today) is AFTER the report run date
	client3ID := suite.seeder.CreateClient(ctx, "Juan", "3", "32345678", "3234", "ACTIVE")
	suite.seeder.CreateInvoice(ctx, client3ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.CreatePaymentForCustomerCreditTest(ctx, "32345678", "UNAPPLIED", today, false, testhelpers.DateHelper{})

	// only consider ledger allocations with a status on unapplied or reapplied
	// this ledger amount will NOT be included in the credit amount on the report as the status is ALLOCATED
	client4ID := suite.seeder.CreateClient(ctx, "Edmond", "4", "52345678", "5234", "ACTIVE")
	suite.seeder.CreateInvoice(ctx, client4ID, shared.InvoiceTypeAD, nil, yesterday.StringPtr(), nil, nil, nil, yesterday.StringPtr())
	suite.CreatePaymentForCustomerCreditTest(ctx, "52345678", "ALLOCATED", yesterday, true, yesterday)

	c := Client{suite.seeder.Conn}
	to := shared.NewDate(yesterday.String())
	rows, err := c.Run(ctx, NewCustomerCredit(CustomerCreditInput{
		ToDate: &to,
	}))

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, len(rows))

	results := mapByHeader(rows)
	assert.NotEmpty(suite.T(), results)

	// client 1
	assert.Equal(suite.T(), "Ianthe 0", results[0]["Customer name"], "Customer name - client 1")
	assert.Equal(suite.T(), "12345679", results[0]["Customer number"], "Customer number - client 1")
	assert.Equal(suite.T(), "1233", results[0]["SOP number"], "SOP number - client 1")
	assert.Equal(suite.T(), "911.11", results[0]["Credit balance"], "Credit balance - client 1")

	// client 2
	assert.Equal(suite.T(), "Beryl 2", results[1]["Customer name"], "Customer name - client 2")
	assert.Equal(suite.T(), "22345678", results[1]["Customer number"], "Customer number - client 2")
	assert.Equal(suite.T(), "2234", results[1]["SOP number"], "SOP number - client 2")
	assert.Equal(suite.T(), "911.11", results[1]["Credit balance"], "Credit balance - client 2")
}

func (suite *IntegrationSuite) CreatePaymentForCustomerCreditTest(ctx context.Context, courtRef string, ledgerStatus string, ledgerDateTime testhelpers.DateHelper, hasCreatedAt bool, ledgerCreatedAt testhelpers.DateHelper) {
	today := suite.seeder.Today()
	twoYearsAgo := today.Sub(2, 0, 0)
	yesterday := today.Sub(0, 0, 1)

	payment := shared.PaymentDetails{
		Amount: 111111,
		BankDate: pgtype.Date{
			Time:  twoYearsAgo.Date(),
			Valid: true,
		},
		CourtRef: pgtype.Text{
			String: courtRef,
			Valid:  true,
		},
		LedgerType: shared.TransactionTypeOPGBACSPayment,
		ReceivedDate: pgtype.Timestamp{
			Time:  yesterday.Date(),
			Valid: true,
		},
		CreatedBy: pgtype.Int4{
			Int32: ctx.(auth.Context).User.ID,
			Valid: true,
		},
		PisNumber: pgtype.Int4{
			Int32: 0,
			Valid: false,
		},
	}

	latestLedgerId := suite.seeder.GetLatestLedgerID(ctx)

	tx, err := suite.seeder.Service.BeginStoreTx(ctx)
	assert.NoError(suite.T(), err, "failed to begin transaction: %v", err)

	_, err = suite.seeder.Service.ProcessPaymentsUploadLine(ctx, tx, payment)
	assert.NoError(suite.T(), err, "payment not processed: %v", err)

	err = tx.Commit(ctx)
	assert.NoError(suite.T(), err, "failed to commit payment: %v", err)

	if hasCreatedAt == false {
		_, err = suite.seeder.Conn.Exec(ctx, "UPDATE supervision_finance.ledger SET datetime = $1, created_at = NULL WHERE id > $2", ledgerDateTime.Date(), latestLedgerId)
		assert.NoError(suite.T(), err, "failed to update ledger dates for refund: %v", err)
	} else {
		_, err = suite.seeder.Conn.Exec(ctx, "UPDATE supervision_finance.ledger SET datetime = $1, created_at = $2 WHERE id > $3", ledgerDateTime.Date(), ledgerCreatedAt.Date(), latestLedgerId)
		assert.NoError(suite.T(), err, "failed to update ledger dates for refund: %v", err)
	}
	_, err = suite.seeder.Conn.Exec(ctx, "UPDATE supervision_finance.ledger_allocation SET datetime = $1, status = $2 WHERE ledger_id > $3", ledgerCreatedAt.Date(), ledgerStatus, latestLedgerId)
	assert.NoError(suite.T(), err, "failed to update ledger allocation dates for refund: %v", err)
}
