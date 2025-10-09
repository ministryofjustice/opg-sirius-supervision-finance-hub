package service

import (
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_processAdhocEvent() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (99, 1, 'no-invoice', 'DEMANDED', NULL);",
		"INSERT INTO ledger VALUES (1, '1', '2022-04-02T00:00:00+00:00', '', 8000, 'Overpayment', 'CARD PAYMENT', 'CONFIRMED', 99, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, NULL, '2022-04-02T00:00:00+00:00', -8000, 'UNAPPLIED', NULL, '', '2022-04-02', NULL);",

		"INSERT INTO invoice VALUES (1, 1, 99, 'S2', 'S200001/20', '2020-04-02', '2020-04-02', 10000, NULL, NULL, NULL, '2020-04-02', NULL, NULL, 5000, '2020-04-02', 1);",
		"INSERT INTO ledger VALUES (2, '2', '2020-04-02T00:00:00+00:00', '', 5000, 'half paid', 'CREDIT REMISSION', 'CONFIRMED', 99, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 1, '2020-04-02T00:00:00+00:00', 5000, 'ALLOCATED', NULL, '', '2020-04-02', NULL);",
		"INSERT INTO invoice VALUES (2, 1, 99, 'S2', 'S200002/21', '2021-04-02', '2021-04-02', 10000, NULL, NULL, NULL, '2021-04-02', NULL, NULL, NULL, '2021-04-02', 1);",
		"ALTER SEQUENCE ledger_id_seq RESTART WITH 3;",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 3;",
	)
	dispatch := &mockDispatch{}
	s := Service{store: store.New(seeder.Conn), dispatch: dispatch, tx: seeder.Conn}
	err := s.ProcessAdhocEvent(ctx)
	assert.Nil(suite.T(), err)

	// wait for async process
	time.Sleep(1 * time.Second)

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
