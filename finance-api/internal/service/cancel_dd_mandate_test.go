package service

import (
	"errors"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_CancelDirectDebitMandate() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
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

	err := s.CancelDirectDebitMandate(ctx, 11, shared.CancelMandate{
		Surname:  "Nameson",
		CourtRef: "1234567T",
	})
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "CancelMandate", allpayMock.called[0])

	rows := seeder.QueryRow(ctx, "SELECT payment_method FROM supervision_finance.finance_client WHERE id = 1")
	var paymentMethod string
	_ = rows.Scan(&paymentMethod)

	assert.Equal(suite.T(), "DEMANDED", paymentMethod)
	assert.Equal(suite.T(), event.PaymentMethod{
		ClientID:      11,
		PaymentMethod: shared.PaymentMethodDemanded,
	}, dispatchMock.event)
}

func (suite *IntegrationSuite) TestService_CancelDirectDebitMandate_fails() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
	)

	Store := store.New(seeder.Conn)
	allpayMock := mockAllpay{
		errs: map[string]error{"CancelMandate": errors.New("some error")},
	}
	dispatchMock := mockDispatch{}

	s := &Service{
		store:    Store,
		allpay:   &allpayMock,
		dispatch: &dispatchMock,
		tx:       seeder.Conn,
	}

	err := s.CancelDirectDebitMandate(ctx, 11, shared.CancelMandate{
		Surname:  "Nameson",
		CourtRef: "1234567T",
	})
	assert.Error(suite.T(), err)

	assert.Equal(suite.T(), "CancelMandate", allpayMock.called[0])

	rows := seeder.QueryRow(ctx, "SELECT payment_method FROM supervision_finance.finance_client WHERE id = 1")
	var paymentMethod string
	_ = rows.Scan(&paymentMethod)

	assert.Equal(suite.T(), "DIRECT DEBIT", paymentMethod) // not changed when unable to update in allpay
	assert.Nil(suite.T(), dispatchMock.event)              // event should not have been sent
}
