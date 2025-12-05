package service

import (
	"errors"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_CreateDirectDebitMandate() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DEMANDED', NULL, '1234567T');",
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

	err := s.CreateDirectDebitMandate(ctx, 11, shared.CreateMandate{
		AllPayCustomer: shared.AllPayCustomer{
			ClientReference: "11111111",
			Surname:         "Holder",
		},
		Address: shared.Address{
			Line1:    "1 Main Street",
			Town:     "Mainopolis",
			PostCode: "MP1 2PM",
		},
		BankAccount: struct {
			BankDetails shared.AllPayBankDetails `json:"bankDetails"`
		}{
			BankDetails: shared.AllPayBankDetails{
				AccountName:   "Mrs Account Holder",
				SortCode:      "30-33-30",
				AccountNumber: "12345678",
			},
		},
	})
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), allpayMock.called, []string{"ModulusCheck", "CreateMandate"})

	rows := seeder.QueryRow(ctx, "SELECT payment_method FROM supervision_finance.finance_client WHERE id = 1")
	var paymentMethod string
	_ = rows.Scan(&paymentMethod)

	assert.Equal(suite.T(), "DIRECT DEBIT", paymentMethod)
	assert.Equal(suite.T(), event.PaymentMethod{
		ClientID:      11,
		PaymentMethod: shared.PaymentMethodDirectDebit,
	}, dispatchMock.event)

	rows = seeder.QueryRow(ctx, "SELECT type FROM supervision_finance.payment_method WHERE finance_client_id = 1")
	var paymentMethodDbEntry string
	_ = rows.Scan(&paymentMethodDbEntry)
	assert.Equal(suite.T(), "DIRECT DEBIT", paymentMethodDbEntry)

}

func (suite *IntegrationSuite) TestService_CreateDirectDebitMandate_modulusCheckFails() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DEMANDED', NULL, '1234567T');",
	)

	Store := store.New(seeder.Conn)
	allpayMock := mockAllpay{
		errs: map[string]error{"ModulusCheck": errors.New("some error")},
	}
	dispatchMock := mockDispatch{}

	s := &Service{
		store:    Store,
		allpay:   &allpayMock,
		dispatch: &dispatchMock,
		tx:       seeder.Conn,
	}

	err := s.CreateDirectDebitMandate(ctx, 11, shared.CreateMandate{
		AllPayCustomer: shared.AllPayCustomer{
			ClientReference: "11111111",
			Surname:         "Holder",
		},
		Address: shared.Address{
			Line1:    "1 Main Street",
			Town:     "Mainopolis",
			PostCode: "MP1 2PM",
		},
		BankAccount: struct {
			BankDetails shared.AllPayBankDetails `json:"bankDetails"`
		}{
			BankDetails: shared.AllPayBankDetails{
				AccountName:   "Mrs Account Holder",
				SortCode:      "30-33-30",
				AccountNumber: "12345678",
			},
		},
	})

	var e *apierror.BadRequest
	if errors.As(err, &e) {
		assert.Equal(suite.T(), "ModulusCheck", e.Field)
		assert.Equal(suite.T(), "Failed", e.Reason)
	} else {
		suite.T().Error("error is not of type BadRequest")
	}

	rows := seeder.QueryRow(ctx, "SELECT payment_method FROM supervision_finance.finance_client WHERE id = 1")
	var paymentMethod string
	_ = rows.Scan(&paymentMethod)

	assert.Equal(suite.T(), "DEMANDED", paymentMethod) // not changed when unable to update in allpay
	assert.Nil(suite.T(), dispatchMock.event)          // event should not have been sent
}

func (suite *IntegrationSuite) TestService_CreateDirectDebitMandate_createMandateFails() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DEMANDED', NULL, '1234567T');",
	)

	Store := store.New(seeder.Conn)
	allpayMock := mockAllpay{
		errs: map[string]error{"CreateMandate": errors.New("some error")},
	}
	dispatchMock := mockDispatch{}

	s := &Service{
		store:    Store,
		allpay:   &allpayMock,
		dispatch: &dispatchMock,
		tx:       seeder.Conn,
	}

	err := s.CreateDirectDebitMandate(ctx, 11, shared.CreateMandate{
		AllPayCustomer: shared.AllPayCustomer{
			ClientReference: "11111111",
			Surname:         "Holder",
		},
		Address: shared.Address{
			Line1:    "1 Main Street",
			Town:     "Mainopolis",
			PostCode: "MP1 2PM",
		},
		BankAccount: struct {
			BankDetails shared.AllPayBankDetails `json:"bankDetails"`
		}{
			BankDetails: shared.AllPayBankDetails{
				AccountName:   "Mrs Account Holder",
				SortCode:      "30-33-30",
				AccountNumber: "12345678",
			},
		},
	})
	var e *apierror.BadRequest
	if errors.As(err, &e) {
		assert.Equal(suite.T(), "Allpay", e.Field)
		assert.Equal(suite.T(), "Failed", e.Reason)
	} else {
		suite.T().Error("error is not of type BadRequest")
	}
	assert.EqualValues(suite.T(), allpayMock.called, []string{"ModulusCheck", "CreateMandate"})

	rows := seeder.QueryRow(ctx, "SELECT payment_method FROM supervision_finance.finance_client WHERE id = 1")
	var paymentMethod string
	_ = rows.Scan(&paymentMethod)

	assert.Equal(suite.T(), "DEMANDED", paymentMethod) // not changed when unable to update in allpay
	assert.Nil(suite.T(), dispatchMock.event)          // event should not have been sent
}
