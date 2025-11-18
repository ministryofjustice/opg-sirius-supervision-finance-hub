package service

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_CheckPaymentMethod() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DEMANDED', NULL, '12345678');",
		"INSERT INTO finance_client VALUES (2, 22, '4321', 'DIRECT DEBIT', NULL, '87654321');",
	)

	Store := store.New(seeder.Conn)
	allpayMock := mockAllpay{}
	dispatchMock := mockDispatch{}

	s := &Service{
		store:    Store,
		allpay:   &allpayMock,
		dispatch: &dispatchMock,
		tx:       seeder.Conn,
	}

	tests := []struct {
		clientID       int32
		expectedCalled bool
	}{
		{clientID: 11, expectedCalled: false},
		{clientID: 22, expectedCalled: true},
	}

	for _, tt := range tests {
		err := s.CheckPaymentMethod(ctx, tt.clientID)
		assert.NoError(suite.T(), err)
		if tt.expectedCalled {
			assert.Equal(suite.T(), tt.clientID, dispatchMock.event.(event.DirectDebitMandateReview).ClientID)
		} else {
			assert.Nil(suite.T(), dispatchMock.event)
		}
	}
}
