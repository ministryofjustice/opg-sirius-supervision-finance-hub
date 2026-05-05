package service

import (
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_processAdhocEvent_unknownTask() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	s := Service{store: store.New(seeder.Conn), tx: seeder.Conn}

	err := s.ProcessAdhocEvent(ctx, shared.AdhocEvent{Task: "unknown"})
	assert.ErrorContains(suite.T(), err, "invalid adhoc process: unknown")
}

func (suite *IntegrationSuite) Test_processAdhocEvent_changePendingCollectionDate() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'no-invoice', 'DEMANDED', NULL);",
		"INSERT INTO pending_collection VALUES (1, 1, '2024-01-01', 8000, 'PENDING', NULL, '2024-01-01', 1);",
		"INSERT INTO pending_collection VALUES (2, 1, '2024-01-01', 8000, 'COLLECTED', NULL, '2024-01-01', 1);",
		"INSERT INTO pending_collection VALUES (3, 1, '2024-01-01', 8000, 'CANCELLED', NULL, '2024-01-01', 1);",
	)
	dispatch := &mockDispatch{}
	s := Service{store: store.New(seeder.Conn), dispatch: dispatch, tx: seeder.Conn}

	var count int

	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection;").Scan(&count)
	assert.Equal(suite.T(), 3, count)

	err := s.ProcessAdhocEvent(ctx, shared.AdhocEvent{Task: "ChangePendingCollectionDate"})
	assert.Nil(suite.T(), err)

	// wait for async process
	time.Sleep(1 * time.Second)

	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection WHERE collection_date = '2026-05-26';").Scan(&count)
	assert.Equal(suite.T(), 1, count)
}
