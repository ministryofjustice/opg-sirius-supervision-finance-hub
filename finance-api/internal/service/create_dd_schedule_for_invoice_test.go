package service

import (
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_CreateDirectDebitScheduleForInvoice() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL);",
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 10000, NULL, '2024-01-01', NULL, '2024-01-01')",
		"INSERT INTO invoice VALUES (2, 11, 1, 'S2', 'S200124/24', '2024-01-01', '2025-03-31', 1000, NULL, '2024-01-01', NULL, '2024-01-01')",
	)

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{}
	dispatchMock := &mockDispatch{}

	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn, dispatch: dispatchMock}

	err := s.CreateDirectDebitScheduleForInvoice(ctx, 11, shared.CreateScheduleForInvoice{
		CreateSchedule: shared.CreateSchedule{
			AllPayCustomer: shared.AllPayCustomer{
				Surname:         "Scheduleson",
				ClientReference: "1234567T",
			},
		},
		InvoiceId: 1,
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

	expected := store.PendingCollection{
		ID:              1,
		FinanceClientID: pgtype.Int4{Int32: 1, Valid: true},
		Amount:          11000,
		CollectionDate:  pgtype.Date{Time: govUKMock.WorkingDay, Valid: true},
	}

	assert.EqualValues(suite.T(), expected, p)

	expectedEvent := event.DirectDebitCollection{
		ClientID:       11,
		Amount:         11000,
		CollectionDate: govUKMock.WorkingDay,
	}
	assert.Equal(suite.T(), expectedEvent, dispatchMock.event)
}

func (suite *IntegrationSuite) TestService_CreateDirectDebitScheduleForInvoice_invoiceHasNoBalance() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL);",
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 10000, NULL, '2024-01-01', NULL, '2024-01-01')",
		"INSERT INTO ledger VALUES (1, '1', '2020-04-02T00:00:00+00:00', '', 10000, 'Settled', 'CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2020-04-02T00:00:00+00:00', 10000, 'ALLOCATED', NULL, '', '2020-04-02', NULL);",
	)

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{}
	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn}

	err := s.CreateDirectDebitScheduleForInvoice(ctx, 11, shared.CreateScheduleForInvoice{
		CreateSchedule: shared.CreateSchedule{
			AllPayCustomer: shared.AllPayCustomer{
				Surname:         "Scheduleson",
				ClientReference: "1234567T",
			},
		},
		InvoiceId: 1,
	})

	assert.Nil(suite.T(), err)

	var c int
	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection").Scan(&c)
	assert.Equal(suite.T(), 0, c)
	assert.Len(suite.T(), allPayMock.called, 0)
}

func (suite *IntegrationSuite) TestService_CreateDirectDebitScheduleForInvoice_scheduleAlreadyExists() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	timeNow := time.Now()
	addWorkingDaysResult := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day()+14, 0, 0, 0, 0, time.UTC)
	nextWorkingDayAfterResult := time.Date(timeNow.Year(), timeNow.Month()+1, 24, 0, 0, 0, 0, time.UTC)
	if addWorkingDaysResult.Day() >= 23 {
		nextWorkingDayAfterResult = time.Date(timeNow.Year(), timeNow.Month()+2, 24, 0, 0, 0, 0, time.UTC)
	}
	collectionDate := nextWorkingDayAfterResult.AddDate(0, 0, 1).Format("2006-01-02")

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{NonWorkingDays: []time.Time{
		addWorkingDaysResult,
		nextWorkingDayAfterResult,
	}}

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL);",
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 10000, NULL, '2024-01-01', NULL, '2024-01-01')",
		fmt.Sprintf("INSERT INTO pending_collection VALUES (1, 1, '%s', 5000, 'PENDING', null, '2024-01-01 00:00:00', 1)", collectionDate),
	)

	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn}

	err := s.CreateDirectDebitScheduleForInvoice(ctx, 11, shared.CreateScheduleForInvoice{
		CreateSchedule: shared.CreateSchedule{
			AllPayCustomer: shared.AllPayCustomer{
				Surname:         "Scheduleson",
				ClientReference: "1234567T",
			},
		},
		InvoiceId: 1,
	})

	assert.Nil(suite.T(), err)

	var c int
	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection").Scan(&c)
	assert.Equal(suite.T(), 1, c)
	assert.Len(suite.T(), allPayMock.called, 0)
}
