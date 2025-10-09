package service

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"time"
)

func (suite *IntegrationSuite) TestService_CancelFeeReduction() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (33, 33, '1234', 'DEMANDED', NULL);",
		"INSERT INTO fee_reduction VALUES (33, 33, 'REMISSION', NULL, '2019-04-01', '2021-03-31', 'Remission to see the notes', FALSE, '2019-05-01');",
	)

	Store := store.New(seeder.Conn)

	s := &Service{
		store: Store,
		tx:    seeder.Conn,
	}

	err := s.CancelFeeReduction(ctx, 33, shared.CancelFeeReduction{CancellationReason: "Reason for cancellation"})
	rows := seeder.QueryRow(ctx, "SELECT deleted, cancelled_at, cancelled_by, cancellation_reason FROM supervision_finance.fee_reduction WHERE id = 33")

	var (
		deleted            bool
		cancelledAt        time.Time
		cancelledBy        int
		cancellationReason string
	)

	_ = rows.Scan(&deleted, &cancelledAt, &cancelledBy, &cancellationReason)

	assert.Equal(suite.T(), true, deleted)
	assert.Equal(suite.T(), 10, cancelledBy)
	assert.NotNil(suite.T(), cancelledAt)
	assert.Equal(suite.T(), "Reason for cancellation", cancellationReason)

	if err == nil {
		return
	}
	suite.T().Error("Cancel fee reduction failed")
}
