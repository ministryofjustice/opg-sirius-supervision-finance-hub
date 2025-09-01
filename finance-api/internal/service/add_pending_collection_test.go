package service

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_AddPendingCollection() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 2, '1234', 'DEMANDED', NULL);",
	)

	s := NewService(seeder.Conn, nil, nil, nil, nil)

	err := s.AddPendingCollection(ctx, 2, shared.PendingCollection{
		Amount:         52000,
		CollectionDate: shared.NewDate("2022-04-02"),
	})

	assert.Nil(suite.T(), err)

	var p store.PendingCollection
	q := seeder.QueryRow(ctx, "SELECT id, finance_client_id, amount, collection_date FROM pending_collection LIMIT 1")
	_ = q.Scan(
		&p.ID,
		&p.FinanceClientID,
		&p.Amount,
		&p.CollectionDate,
	)

	date, _ := time.Parse("2006-01-02", "2022-04-02")
	expected := store.PendingCollection{
		ID:              1,
		FinanceClientID: pgtype.Int4{Int32: 1, Valid: true},
		Amount:          52000,
		CollectionDate:  pgtype.Date{Time: date, Valid: true},
	}

	assert.EqualValues(suite.T(), expected, p)
}
