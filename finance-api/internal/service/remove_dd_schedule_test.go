package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (suite *IntegrationSuite) TestService_RemoveDirectDebitSchedule() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	collectionDate := time.Now().AddDate(0, 0, 10).UTC().Truncate(24 * time.Hour)

	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Person', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'Client', NULL);",
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
		fmt.Sprintf("INSERT INTO pending_collection VALUES (1, 1, '%s', 12300, 'PENDING', NULL, '2025-10-10', 1)", collectionDate.Format("2006-01-02")),
	)

	Store := store.New(seeder.Conn)
	allpayMock := mockAllpay{}

	s := &Service{
		store:  Store,
		allpay: &allpayMock,
		tx:     seeder.Conn,
		env:    &Env{AllpayEnabled: true},
	}

	err := s.RemoveDirectDebitSchedule(ctx, shared.RemoveSchedule{
		AllPayCustomer: shared.AllPayCustomer{
			ClientReference: "1234567T",
			Surname:         "Person",
		},
		CollectionDate: collectionDate,
		Amount:         12300,
	})

	suite.NoError(err)
	suite.Equal("RemoveSchedule", allpayMock.called[0])

	rows := seeder.QueryRow(ctx, "SELECT status FROM supervision_finance.pending_collection WHERE id = 1")
	var status string
	_ = rows.Scan(&status)
	suite.Equal("CANCELLED", status)
}

func (suite *IntegrationSuite) TestService_RemoveDirectDebitSchedule_fails() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	collectionDate := time.Now().AddDate(0, 0, 10).UTC().Truncate(24 * time.Hour)

	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Person', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'Client', NULL);",
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
		fmt.Sprintf("INSERT INTO pending_collection VALUES (1, 1, '%s', 12300, 'PENDING', NULL, '2025-10-10', 1)", collectionDate.Format("2006-01-02")),
	)

	Store := store.New(seeder.Conn)
	allpayMock := mockAllpay{
		errs: map[string]error{"RemoveSchedule": errors.New("some error")},
	}

	s := &Service{
		store:  Store,
		allpay: &allpayMock,
		tx:     seeder.Conn,
		env:    &Env{AllpayEnabled: true},
	}

	err := s.RemoveDirectDebitSchedule(ctx, shared.RemoveSchedule{
		AllPayCustomer: shared.AllPayCustomer{
			ClientReference: "1234567T",
			Surname:         "Person",
		},
		CollectionDate: collectionDate,
		Amount:         12300,
	})

	suite.Error(err)
	suite.Equal("RemoveSchedule", allpayMock.called[0])

	// pending collection should NOT be cancelled when allpay fails
	rows := seeder.QueryRow(ctx, "SELECT status FROM supervision_finance.pending_collection WHERE id = 1")
	var status string
	_ = rows.Scan(&status)
	suite.Equal("PENDING", status)
}

func (suite *IntegrationSuite) TestService_RemoveDirectDebitSchedule_no_matching_pending_collection() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	collectionDate := time.Now().AddDate(0, 0, 10).UTC().Truncate(24 * time.Hour)

	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Person', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'Client', NULL);",
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
	)

	Store := store.New(seeder.Conn)
	allpayMock := mockAllpay{}

	s := &Service{
		store:  Store,
		allpay: &allpayMock,
		tx:     seeder.Conn,
		env:    &Env{AllpayEnabled: true},
	}

	// should succeed even if there is no matching pending collection to cancel
	err := s.RemoveDirectDebitSchedule(ctx, shared.RemoveSchedule{
		AllPayCustomer: shared.AllPayCustomer{
			ClientReference: "1234567T",
			Surname:         "Person",
		},
		CollectionDate: collectionDate,
		Amount:         12300,
	})

	suite.NoError(err)
	suite.Equal("RemoveSchedule", allpayMock.called[0])
}


