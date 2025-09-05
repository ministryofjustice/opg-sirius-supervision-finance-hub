package service

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_expireRefunds() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	today := time.Now()
	aWeekAgo := today.AddDate(0, 0, -7).Format("2006-01-02")
	tooOld := today.AddDate(0, 0, -15).Format("2006-01-02")

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'ian-test', 'DEMANDED', NULL, '12345678');",

		"INSERT INTO refund VALUES (1, 1, '2019-01-05', 32000, 'PENDING', 'An old pending refund', 99, '"+tooOld+"')",
		"INSERT INTO refund VALUES (2, 1, '2019-01-05', 32000, 'PENDING', 'An new pending refund', 99, '"+aWeekAgo+"')",
		"INSERT INTO refund VALUES (3, 1, '2019-01-05', 32000, 'APPROVED', 'An old processing refund', 99, '"+tooOld+"', 99, '"+tooOld+"', '"+tooOld+"')",
		"INSERT INTO refund VALUES (4, 1, '2019-01-05', 32000, 'APPROVED', 'A new processing refund', 99, '"+tooOld+"', 99, '"+tooOld+"', '"+aWeekAgo+"')",
		"INSERT INTO refund VALUES (5, 1, '2019-01-06', 15500, 'APPROVED', 'An old approved refund', 99, '"+tooOld+"', 99, '"+tooOld+"')",
		"INSERT INTO refund VALUES (6, 1, '2019-01-06', 15500, 'APPROVED', 'A new approved refund', 99, '"+tooOld+"', 99, '"+aWeekAgo+"')",

		"INSERT INTO bank_details VALUES (1, 1, 'MR IAN TEST', '11111111', '11-11-11');",
		"INSERT INTO bank_details VALUES (2, 2, 'MR IAN TEST', '11111111', '11-11-11');",
		"INSERT INTO bank_details VALUES (3, 3, 'MR IAN TEST', '11111111', '11-11-11');",
		"INSERT INTO bank_details VALUES (4, 4, 'MR IAN TEST', '11111111', '11-11-11');",
		"INSERT INTO bank_details VALUES (5, 5, 'MR IAN TEST', '11111111', '11-11-11');",
		"INSERT INTO bank_details VALUES (6, 6, 'MR IAN TEST', '11111111', '11-11-11');",
	)

	s := Service{store: store.New(seeder.Conn)}

	suite.T().Run("ExpireRefunds", func(t *testing.T) {
		err := s.ExpireRefunds(ctx)
		assert.NoError(t, err)

		var count int
		_ = seeder.QueryRow(suite.ctx, `SELECT COUNT(id) FROM refund WHERE decision = 'REJECTED'`).Scan(&count)
		assert.Equal(t, 1, count)

		_ = seeder.QueryRow(suite.ctx, `SELECT COUNT(id) FROM refund WHERE cancelled_at IS NOT NULL`).Scan(&count)
		assert.Equal(t, 2, count)

		_ = seeder.QueryRow(suite.ctx, `SELECT COUNT(id) FROM bank_details`).Scan(&count)
		assert.Equal(t, 3, count)
	})
}
