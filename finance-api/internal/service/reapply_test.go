package service

import (
	"context"
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
	invoiceBalanceForQuery = `SELECT SUM(la.amount) FROM supervision_finance.ledger_allocation la WHERE la.invoice_id = $1;`
)

func (suite *IntegrationSuite) TestService_reapplyCredit_noInvoices() {
	conn := suite.testDB.GetConn()
	ctx := context.Background()

	conn.SeedData(
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
	s := NewService(conn.Conn)
	err := s.reapplyCredit(ctx, 1)
	assert.Nil(suite.T(), err)

	var credit int
	row := conn.QueryRow(ctx, customerCreditQuery)
	_ = row.Scan(&credit)

	assert.Equal(suite.T(), 10000, credit)
}

func (suite *IntegrationSuite) TestService_reapplyCredit_oldestFirst() {
	conn := suite.testDB.GetConn()
	ctx := context.Background()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'no-invoice', 'DEMANDED', NULL);",
		"INSERT INTO ledger VALUES (1, '1', '2022-04-02T00:00:00+00:00', '', 8000, 'Overpayment', 'CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, NULL, '2022-04-02T00:00:00+00:00', -8000, 'UNAPPLIED', NULL, '', '2022-04-02', NULL);",

		// two invoices partially paid
		"INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S200001/20', '2020-04-02', '2020-04-02', 10000, NULL, NULL, NULL, '2020-04-02', NULL, NULL, 0, '2020-04-02', 1);",
		"INSERT INTO ledger VALUES (2, '2', '2020-04-02T00:00:00+00:00', '', 5000, 'half paid', 'CREDIT REMISSION', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 1, '2020-04-02T00:00:00+00:00', 5000, 'ALLOCATED', NULL, '', '2020-04-02', NULL);",
		"INSERT INTO invoice VALUES (2, 1, 1, 'S2', 'S200002/21', '2021-04-02', '2021-04-02', 10000, NULL, NULL, NULL, '2021-04-02', NULL, NULL, 0, '2021-04-02', 1);",
		"ALTER SEQUENCE ledger_id_seq RESTART WITH 3;",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 3;",
	)
	s := NewService(conn.Conn)
	err := s.reapplyCredit(ctx, 1)
	assert.Nil(suite.T(), err)

	var amount int
	row := conn.QueryRow(ctx, customerCreditQuery)
	_ = row.Scan(&amount)

	assert.Equal(suite.T(), 0, amount)

	row = conn.QueryRow(ctx, countReappliedQuery)
	_ = row.Scan(&amount)

	// two new reapply allocations
	assert.Equal(suite.T(), 2, amount)

	_, _ = conn.Prepare(ctx, "invoice_balance", invoiceBalanceForQuery)
	row = conn.QueryRow(ctx, "invoice_balance", 1)
	_ = row.Scan(&amount)

	// pays off the oldest in full
	assert.Equal(suite.T(), 10000, amount)

	row = conn.QueryRow(ctx, "invoice_balance", 2)
	_ = row.Scan(&amount)

	// the remainder goes to the next oldest
	assert.Equal(suite.T(), 3000, amount)
}
