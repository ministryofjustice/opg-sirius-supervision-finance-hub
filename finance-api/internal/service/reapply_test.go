package service

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/stretchr/testify/assert"
)

const (
	customerCreditQuery = `SELECT ABS(SUM(la.amount)) FROM finance_client fc 
								JOIN supervision_finance.ledger l ON fc.id = l.finance_client_id 
								JOIN supervision_finance.ledger_allocation la ON l.id = la.ledger_id 
                      				WHERE la.status IN ('UNAPPLIED', 'REAPPLIED') AND fc.client_id = 1;`
	countReappliedQuery = `SELECT COUNT(*) FROM finance_client fc 
								JOIN supervision_finance.ledger l ON fc.id = l.finance_client_id 
								JOIN supervision_finance.ledger_allocation la ON l.id = la.ledger_id 
                      				WHERE la.status = 'REAPPLIED' AND fc.client_id = 1;`
)

func (suite *IntegrationSuite) TestService_reapplyCredit_noInvoices() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'no-invoice', 'DEMANDED', NULL);",
		"INSERT INTO ledger VALUES (1, '1', '2022-04-02T00:00:00+00:00', '', -10000, 'Overpayment', 'CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, NULL, '2022-04-02T00:00:00+00:00', -10000, 'UNAPPLIED', NULL, '', '2022-04-02', NULL);",

		// only invoice is settled
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S200001/20', '2020-04-02', '2020-04-02', 10000, NULL, NULL, NULL, '2020-04-02', NULL, NULL, 0, '2020-04-02', 1);",
		"INSERT INTO ledger VALUES (2, '2', '2020-04-02T00:00:00+00:00', '', 10000, 'Settled', 'CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 1, '2020-04-02T00:00:00+00:00', 10000, 'ALLOCATED', NULL, '', '2020-04-02', NULL);",
		"ALTER SEQUENCE ledger_id_seq RESTART WITH 3;",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 3;",
	)
	dispatch := &mockDispatch{}
	s := NewService(seeder.Conn, dispatch, nil, nil, nil)
	err := s.ReapplyCredit(ctx, 1, nil)
	assert.Nil(suite.T(), err)

	var credit int
	row := seeder.QueryRow(ctx, customerCreditQuery)
	_ = row.Scan(&credit)

	assert.Equal(suite.T(), 10000, credit)

	expected := event.CreditOnAccount{
		ClientID:        1,
		CreditRemaining: 10000,
	}
	assert.Equal(suite.T(), expected, dispatch.event)
}

func (suite *IntegrationSuite) TestService_reapplyCredit_oldestFirst() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'no-invoice', 'DEMANDED', NULL);",
		"INSERT INTO ledger VALUES (1, '1', '2022-04-02T00:00:00+00:00', '', 8000, 'Overpayment', 'CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, NULL, '2022-04-02T00:00:00+00:00', -8000, 'UNAPPLIED', NULL, '', '2022-04-02', NULL);",

		// two invoices partially paid
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S200001/20', '2020-04-02', '2020-04-02', 10000, NULL, NULL, NULL, '2020-04-02', NULL, NULL, 5000, '2020-04-02', 1);",
		"INSERT INTO ledger VALUES (2, '2', '2020-04-02T00:00:00+00:00', '', 5000, 'half paid', 'CREDIT REMISSION', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 1, '2020-04-02T00:00:00+00:00', 5000, 'ALLOCATED', NULL, '', '2020-04-02', NULL);",
		"INSERT INTO invoice VALUES (2, 1, 1, 'S2', 'S200002/21', '2021-04-02', '2021-04-02', 10000, NULL, NULL, NULL, '2021-04-02', NULL, NULL, NULL, '2021-04-02', 1);",
		"ALTER SEQUENCE ledger_id_seq RESTART WITH 3;",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 3;",
	)
	dispatch := &mockDispatch{}
	s := NewService(seeder.Conn, dispatch, nil, nil, nil)
	err := s.ReapplyCredit(ctx, 1, nil)
	assert.Nil(suite.T(), err)

	var (
		amount           int
		cachedDebtAmount int
	)
	row := seeder.QueryRow(ctx, customerCreditQuery)
	_ = row.Scan(&amount)

	assert.Equal(suite.T(), 0, amount)

	row = seeder.QueryRow(ctx, countReappliedQuery)
	_ = row.Scan(&amount)

	// two new reapply allocations
	assert.Equal(suite.T(), 2, amount)

	row = seeder.QueryRow(ctx, "SELECT SUM(la.amount), i.cacheddebtamount FROM supervision_finance.ledger_allocation la JOIN supervision_finance.invoice i ON i.id = la.invoice_id WHERE la.invoice_id = 1 GROUP BY i.cacheddebtamount;")
	_ = row.Scan(&amount, &cachedDebtAmount)

	// pays off the oldest in full
	assert.Equal(suite.T(), 10000, amount)
	assert.Equal(suite.T(), 0, cachedDebtAmount)

	row = seeder.QueryRow(ctx, "SELECT SUM(la.amount), i.cacheddebtamount FROM supervision_finance.ledger_allocation la JOIN supervision_finance.invoice i ON i.id = la.invoice_id WHERE la.invoice_id = 2 GROUP BY i.cacheddebtamount;")
	_ = row.Scan(&amount, &cachedDebtAmount)

	// the remainder goes to the next oldest
	assert.Equal(suite.T(), 3000, amount)
	assert.Equal(suite.T(), 7000, cachedDebtAmount)

	assert.Nil(suite.T(), dispatch.event)
}

func (suite *IntegrationSuite) TestService_reapplyCredit_requiresApprovedLedger() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'no-invoice', 'DEMANDED', NULL);",
		"INSERT INTO ledger VALUES (1, '1', '2022-04-02T00:00:00+00:00', '', 15000, 'Overpayment', 'CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, NULL, '2022-04-02T00:00:00+00:00', -15000, 'UNAPPLIED', NULL, '', '2022-04-02', NULL);",

		// invoice partially paid with approved (not confirmed) credit memo
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S200001/20', '2020-04-02', '2020-04-02', 10000, NULL, NULL, NULL, '2020-04-02', NULL, NULL, 5000, '2020-04-02', 1);",
		"INSERT INTO ledger VALUES (2, '2', '2020-04-02T00:00:00+00:00', '', 5000, 'half paid', 'CREDIT MEMO', 'APPROVED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 1, '2020-04-02T00:00:00+00:00', 5000, 'ALLOCATED', NULL, '', '2020-04-02', NULL);",
		"ALTER SEQUENCE ledger_id_seq RESTART WITH 3;",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 3;",
	)
	dispatch := &mockDispatch{}
	s := NewService(seeder.Conn, dispatch, nil, nil, nil)
	err := s.ReapplyCredit(ctx, 1, nil)
	assert.Nil(suite.T(), err)

	var (
		amount           int
		cachedDebtAmount int
	)
	row := seeder.QueryRow(ctx, customerCreditQuery)
	_ = row.Scan(&amount)

	assert.Equal(suite.T(), 5000, amount)

	row = seeder.QueryRow(ctx, countReappliedQuery)
	_ = row.Scan(&amount)

	assert.Equal(suite.T(), 1, amount)

	row = seeder.QueryRow(ctx, "SELECT SUM(la.amount), i.cacheddebtamount FROM supervision_finance.ledger_allocation la JOIN supervision_finance.invoice i ON i.id = la.invoice_id JOIN supervision_finance.ledger l ON la.ledger_id = l.id WHERE la.invoice_id = 1 AND l.status != 'APPROVED' GROUP BY i.cacheddebtamount;")
	_ = row.Scan(&amount, &cachedDebtAmount)

	assert.Equal(suite.T(), 10000, amount)
	assert.Equal(suite.T(), -5000, cachedDebtAmount)
}
