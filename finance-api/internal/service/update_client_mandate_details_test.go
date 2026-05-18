package service

import (
	"errors"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_UpdateClientMandateDetails() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Smith', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'Client', NULL);",
		`INSERT INTO public.addresses VALUES (1, 11, '["1 Test Street"]', 'Testtown', NULL, 'TE1 1ST', NULL, NULL, NULL, NULL);`,
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
	)

	Store := store.New(seeder.Conn)
	allpayMock := mockAllpay{}

	s := &Service{
		store:  Store,
		allpay: &allpayMock,
		tx:     seeder.Conn,
	}

	err := s.UpdateClientMandateDetails(ctx, 11, shared.ClientUpdatedEvent{
		ClientID: 11,
		Surname:  shared.ClientChanges{Old: "Smith", New: "Jones"},
	})
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), []string{"UpdateClientDetails"}, allpayMock.called)

	input := allpayMock.lastCalledParams[0].(*allpay.UpdateClientDetailsInput)
	assert.Equal(suite.T(), "1234567T", input.ClientReference)
	assert.Equal(suite.T(), "Smith", input.Surname)
	assert.Equal(suite.T(), "Jones", input.NewSurname)
	assert.Equal(suite.T(), "1 Test Street", input.Address.Line1)
	assert.Equal(suite.T(), "Testtown", input.Address.Town)
	assert.Equal(suite.T(), "TE1 1ST", input.Address.PostCode)
}

func (suite *IntegrationSuite) TestService_UpdateClientMandateDetails_skips_when_not_direct_debit() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Smith', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'Client', NULL);",
		`INSERT INTO public.addresses VALUES (1, 11, '["1 Test Street"]', 'Testtown', NULL, 'TE1 1ST', NULL, NULL, NULL, NULL);`,
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DEMANDED', NULL, '1234567T');",
	)

	Store := store.New(seeder.Conn)
	allpayMock := mockAllpay{}

	s := &Service{
		store:  Store,
		allpay: &allpayMock,
		tx:     seeder.Conn,
	}

	err := s.UpdateClientMandateDetails(ctx, 11, shared.ClientUpdatedEvent{
		ClientID: 11,
		Surname:  shared.ClientChanges{Old: "Smith", New: "Jones"},
	})
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), allpayMock.called)
}

func (suite *IntegrationSuite) TestService_UpdateClientMandateDetails_fails_when_allpay_errors() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Smith', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'Client', NULL);",
		`INSERT INTO public.addresses VALUES (1, 11, '["1 Test Street"]', 'Testtown', NULL, 'TE1 1ST', NULL, NULL, NULL, NULL);`,
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
	)

	Store := store.New(seeder.Conn)
	allpayMock := mockAllpay{
		errs: map[string]error{"UpdateClientDetails": errors.New("some error")},
	}

	s := &Service{
		store:  Store,
		allpay: &allpayMock,
		tx:     seeder.Conn,
	}

	err := s.UpdateClientMandateDetails(ctx, 11, shared.ClientUpdatedEvent{
		ClientID: 11,
		Surname:  shared.ClientChanges{Old: "Smith", New: "Jones"},
	})
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), []string{"UpdateClientDetails"}, allpayMock.called)
}

