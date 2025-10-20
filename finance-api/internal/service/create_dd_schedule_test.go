package service

import (
	"errors"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_CreateDirectDebitSchedule() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL);",
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 10000, NULL, '2024-01-01', NULL, '2024-01-01')",
	)

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{}
	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn}

	pc, err := s.CreateDirectDebitSchedule(ctx, 11, shared.CreateSchedule{
		AllPayCustomer: shared.AllPayCustomer{
			Surname:         "Scheduleson",
			ClientReference: "1234567T",
		},
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

	collectionDate := govUKMock.WorkingDay.Truncate(24 * time.Hour)
	expected := store.PendingCollection{
		ID:              1,
		FinanceClientID: pgtype.Int4{Int32: 1, Valid: true},
		Amount:          10000,
		CollectionDate:  pgtype.Date{Time: collectionDate, Valid: true},
	}

	assert.EqualValues(suite.T(), expected, p)
	assert.Equal(suite.T(), PendingCollection{
		Amount:         10000,
		CollectionDate: collectionDate,
	}, pc)

	assert.Equal(suite.T(), govUKMock.nWorkingDays, 14)
	assert.Equal(suite.T(), govUKMock.Xday, 24)
	assert.NoError(suite.T(), govUKMock.errs["AddWorkingDays"])
	assert.NoError(suite.T(), govUKMock.errs["NextWorkingDayOnOrAfterX"])

	assert.Equal(suite.T(), "CreateSchedule", allPayMock.called[0])
	assert.Equal(suite.T(), &allpay.CreateScheduleInput{
		Date:   govUKMock.WorkingDay.Truncate(24 * time.Hour),
		Amount: 10000,
		ClientDetails: allpay.ClientDetails{
			ClientReference: "1234567T",
			Surname:         "Scheduleson",
		},
	}, allPayMock.lastCalledParams[0])
}

func (suite *IntegrationSuite) TestService_CreateDirectDebitSchedule_pendingBalanceFails() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL);",
	)

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{}
	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn}

	_, err := s.CreateDirectDebitSchedule(ctx, 99, shared.CreateSchedule{
		AllPayCustomer: shared.AllPayCustomer{
			Surname:         "Scheduleson",
			ClientReference: "1234567T",
		},
	})

	assert.Error(suite.T(), err)

	var c int
	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection").Scan(&c)
	assert.Equal(suite.T(), 0, c)
	assert.Len(suite.T(), allPayMock.called, 0)
}

func (suite *IntegrationSuite) TestService_CreateDirectDebitSchedule_noPendingBalance() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL);",
	)

	allpayMock := &mockAllpay{}
	govUKMock := &mockGovUK{}
	s := Service{store: store.New(seeder.Conn), allpay: allpayMock, govUK: govUKMock, tx: seeder.Conn}

	pc, err := s.CreateDirectDebitSchedule(ctx, 11, shared.CreateSchedule{
		AllPayCustomer: shared.AllPayCustomer{
			Surname:         "Scheduleson",
			ClientReference: "1234567T",
		},
	})

	assert.Nil(suite.T(), err) // no error expected as DD mandate can be set up without a debt

	var c int
	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection").Scan(&c)
	assert.Equal(suite.T(), 0, c)
	assert.Len(suite.T(), allpayMock.called, 0)
	assert.Equal(suite.T(), PendingCollection{}, pc)
}

func (suite *IntegrationSuite) TestService_CreateDirectDebitSchedule_workingDayFails() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL);",
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 10000, NULL, '2024-01-01', NULL, '2024-01-01')",
	)

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{errs: map[string]error{"AddWorkingDays": errors.New("AddWorkingDays error")}}
	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn}

	_, err := s.CreateDirectDebitSchedule(ctx, 11, shared.CreateSchedule{
		AllPayCustomer: shared.AllPayCustomer{
			Surname:         "Scheduleson",
			ClientReference: "1234567T",
		},
	})

	assert.Error(suite.T(), err)

	var c int
	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection").Scan(&c)
	assert.Equal(suite.T(), 0, c)
	assert.Len(suite.T(), allPayMock.called, 0)
}

func (suite *IntegrationSuite) TestService_CreateDirectDebitSchedule_createScheduleFails() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL);",
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 10000, NULL, '2024-01-01', NULL, '2024-01-01')",
	)

	allPayMock := &mockAllpay{errs: map[string]error{"CreateSchedule": errors.New("CreateSchedule error")}}
	govUKMock := &mockGovUK{}
	dispatchMock := &mockDispatch{}
	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn, dispatch: dispatchMock}

	_, err := s.CreateDirectDebitSchedule(ctx, 11, shared.CreateSchedule{
		AllPayCustomer: shared.AllPayCustomer{
			Surname:         "Scheduleson",
			ClientReference: "1234567T",
		},
	})

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "CreateSchedule", allPayMock.called[0])
	assert.Equal(suite.T(), 11, dispatchMock.event.(event.DirectDebitScheduleFailed).ClientID)

	var c int
	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection").Scan(&c)
	assert.Equal(suite.T(), 0, c) // pending collection is rolled back on error
}
