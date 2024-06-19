package service

import (
	"context"
	"database/sql"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/stretchr/testify/assert"
	"time"
)

func (suite *IntegrationSuite) TestService_CancelFeeReduction() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"SET SEARCH_PATH to supervision_finance;",
		"INSERT INTO finance_client VALUES (33, 33, '1234', 'DEMANDED', NULL);",
		"INSERT INTO fee_reduction VALUES (33, 33, 'REMISSION', NULL, '2019-04-01', '2021-03-31', 'Remission to see the notes', FALSE, '2019-05-01');",
	)

	ctx := context.Background()
	Store := store.New(conn)

	s := &Service{
		store: Store,
		tx:    conn,
	}

	err := s.CancelFeeReduction(33)
	rows, _ := conn.Query(ctx, "SELECT * FROM supervision_finance.fee_reduction WHERE id = 33")
	defer rows.Close()

	for rows.Next() {
		var (
			id            int
			financeClient int
			feeType       string
			evidenceType  sql.NullString
			startDate     time.Time
			endDate       time.Time
			notes         string
			deleted       bool
			dateReceived  time.Time
		)

		_ = rows.Scan(&id, &financeClient, &feeType, &evidenceType, &startDate, &endDate, &notes, &deleted, &dateReceived)

		assert.Equal(suite.T(), true, deleted)
		assert.Equal(suite.T(), "Remission to see the notes", notes)
	}

	if err == nil {
		return
	}
	suite.T().Error("Cancel fee reduction failed")
}
