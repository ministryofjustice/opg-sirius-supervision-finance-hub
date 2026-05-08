package service

import (
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) Test_queueScheduleRemovals() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DEMANDED', NULL, '321CBA');",
	)

	suite.T().Run("queue schedule removals", func(t *testing.T) {
		records := [][]string{
			{"Client Reference", "Surname", "Amount"},
			{"123ABC", "Felicity", "123.21"}, // client not found
			{"321CBA", "Brian", "321.23"},
		}
		uploadDate := shared.NewDate("2022-04-02")

		dispatch := &mockDispatch{}
		s := Service{store: store.New(seeder.Conn), dispatch: dispatch, tx: seeder.Conn}
		failedRows := s.QueueScheduleRemovals(suite.ctx, records, uploadDate)

		assert.Len(t, failedRows, 1)

		assert.Len(t, dispatch.called, 1)
		lastEvent := dispatch.event.(event.ScheduleToRemove)
		assert.Equal(t, "321CBA", lastEvent.CourtRef)
		assert.Equal(t, "Brian", lastEvent.Surname)
		assert.Equal(t, int32(32123), lastEvent.Amount)
		assert.Equal(t, uploadDate, lastEvent.Date)
	})
}
