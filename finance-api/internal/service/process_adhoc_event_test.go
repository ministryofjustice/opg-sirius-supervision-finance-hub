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
		"INSERT INTO ledger VALUES (1, '1', '2022-04-02T00:00:00+00:00', '', 8000, 'Overpayment', 'ONLINE CARD PAYMENT', 'CONFIRMED', 99, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, NULL, '2022-04-02T00:00:00+00:00', -8000, 'UNAPPLIED', NULL, '', '2022-04-02', NULL);",

		"INSERT INTO ledger VALUES (2, '2', '2020-04-02T00:00:00+00:00', '', 8000, 'half paid', 'REFUND', 'CONFIRMED', 99, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (2, 2, NULL, '2020-04-02T00:00:00+00:00', 8000, 'REAPPLIED', NULL, '', '2020-04-02', NULL);",
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
		amount int
	)

	_ = seeder.QueryRow(ctx, "SELECT l.amount FROM supervision_finance.ledger l WHERE l.id = 1;").Scan(&amount)

	// don't change amount for non-refund ledgers
	assert.Equal(suite.T(), 8000, amount)

	_ = seeder.QueryRow(ctx, "SELECT l.amount FROM supervision_finance.ledger l WHERE l.id = 2;").Scan(&amount)

	// makes refund ledger amount negative
	assert.Equal(suite.T(), -8000, amount)
}
