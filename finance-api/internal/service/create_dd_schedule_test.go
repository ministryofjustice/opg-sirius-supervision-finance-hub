package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

// Tests for generateScheduleData
func (suite *IntegrationSuite) TestService_generateScheduleData_success() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL);",
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 10000, NULL, '2024-01-01', NULL, '2024-01-01')",
	)

	govUKMock := &mockGovUK{}
	s := Service{store: store.New(seeder.Conn), govUK: govUKMock, tx: seeder.Conn}

	pc, err := s.generateScheduleData(ctx, 11)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), PendingCollection{
		Amount:         10000,
		CollectionDate: govUKMock.WorkingDay.Truncate(24 * time.Hour),
	}, pc)
	assert.Equal(suite.T(), 14, govUKMock.nWorkingDays)
	assert.Equal(suite.T(), 24, govUKMock.Xday)
}

func (suite *IntegrationSuite) TestService_generateScheduleData_noBalance() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL);",
	)

	govUKMock := &mockGovUK{}
	s := Service{store: store.New(seeder.Conn), govUK: govUKMock, tx: seeder.Conn}

	pc, err := s.generateScheduleData(ctx, 11)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), PendingCollection{}, pc)
}

func (suite *IntegrationSuite) TestService_generateScheduleData_balanceFetchFails() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL);",
	)

	govUKMock := &mockGovUK{}
	s := Service{store: store.New(seeder.Conn), govUK: govUKMock, tx: seeder.Conn}

	_, err := s.generateScheduleData(ctx, 99)

	assert.Error(suite.T(), err)
}

func (suite *IntegrationSuite) TestService_generateScheduleData_workingDayCalculationFails() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL);",
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 10000, NULL, '2024-01-01', NULL, '2024-01-01')",
	)

	govUKMock := &mockGovUK{errs: map[string]error{"AddWorkingDays": errors.New("AddWorkingDays error")}}
	s := Service{store: store.New(seeder.Conn), govUK: govUKMock, tx: seeder.Conn}

	_, err := s.generateScheduleData(ctx, 11)

	assert.Error(suite.T(), err)
}

// Tests for CreateDirectDebitSchedule (public entry point)

func (suite *IntegrationSuite) TestService_CreateDirectDebitSchedule_skipsNonDirectDebitClient() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'Client', NULL);",
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DEMANDED', NULL, '1234567T');",
		`INSERT INTO public.addresses VALUES (1, 11, '["1 Test Street"]', 'Testtown', NULL, 'TE1 1ST', NULL);`,
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 5000, NULL, '2024-01-01', NULL, '2024-01-01')",
	)

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{}
	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn, env: &Env{AllpayEnabled: true}}

	err := s.CreateDirectDebitSchedule(ctx, shared.InvoiceCreatedEvent{ClientID: 11, InvoiceID: 1, InvoiceType: shared.InvoiceTypeB2})
	assert.Nil(suite.T(), err)

	var c int
	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection").Scan(&c)
	assert.Equal(suite.T(), 0, c)
	assert.Len(suite.T(), allPayMock.called, 0)
}

func (suite *IntegrationSuite) TestService_CreateDirectDebitSchedule_skipsNonDirectDebitInvoiceType() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'Client', NULL);",
		"INSERT INTO finance_client (id, client_id, sop_number, payment_method, batchnumber, court_ref) VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 5000, NULL, '2024-01-01', NULL, '2024-01-01')",
	)

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{}
	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn, env: &Env{AllpayEnabled: true}}

	err := s.CreateDirectDebitSchedule(ctx, shared.InvoiceCreatedEvent{ClientID: 11, InvoiceID: 1, InvoiceType: shared.InvoiceTypeS2})
	assert.Nil(suite.T(), err)

	var c int
	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection").Scan(&c)
	assert.Equal(suite.T(), 0, c)
	assert.Len(suite.T(), allPayMock.called, 0)
}

func (suite *IntegrationSuite) TestService_CreateDirectDebitSchedule_success() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'Client', NULL);",
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
		`INSERT INTO public.addresses VALUES (1, 11, '["1 Test Street"]', 'Testtown', NULL, 'TE1 1ST', NULL);`,
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 10000, NULL, '2024-01-01', NULL, '2024-01-01')",
		"INSERT INTO invoice VALUES (2, 11, 1, 'S2', 'S200124/24', '2024-01-01', '2025-03-31', 1000, NULL, '2024-01-01', NULL, '2024-01-01')",
	)

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{}
	dispatchMock := &mockDispatch{}

	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn, dispatch: dispatchMock, env: &Env{AllpayEnabled: true}}

	err := s.CreateDirectDebitSchedule(ctx, shared.InvoiceCreatedEvent{ClientID: 11, InvoiceID: 1, InvoiceType: shared.InvoiceTypeB2})

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
	assert.Equal(suite.T(), "CreateSchedule", allPayMock.called[0])
	assert.Equal(suite.T(), &allpay.CreateScheduleInput{
		Date:   govUKMock.WorkingDay.Truncate(24 * time.Hour),
		Amount: 11000,
		ClientDetails: allpay.ClientDetails{
			ClientReference: "1234567T",
			Surname:         "Scheduleson",
		},
	}, allPayMock.lastCalledParams[0])
}

func (suite *IntegrationSuite) TestService_CreateDirectDebitSchedule_skipsWhenScheduleExists() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	// Construct non-working days but ensure the calculated collection date is reproducible
	timeNow := time.Now().UTC()
	addWorkingDaysResult := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day()+14, 0, 0, 0, 0, time.UTC)
	nonWorking := []time.Time{addWorkingDaysResult}

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{NonWorkingDays: nonWorking}

	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'Client', NULL);",
		"INSERT INTO finance_client VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
		`INSERT INTO public.addresses VALUES (1, 11, '["1 Test Street"]', 'Testtown', NULL, 'TE1 1ST', NULL);`,
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 10000, NULL, '2024-01-01', NULL, '2024-01-01')",
	)

	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn, env: &Env{AllpayEnabled: true}}

	scheduleDate, _ := s.calculateScheduleCollectionDate(ctx)

	seeder.SeedData(
		fmt.Sprintf("INSERT INTO pending_collection VALUES (1, 1, '%s', 5000, 'PENDING', NULL, '2024-01-01 00:00:00', 1)", scheduleDate.Format("2006-01-02")),
		"ALTER SEQUENCE supervision_finance.pending_collection_id_seq RESTART WITH 2",
	)

	err := s.CreateDirectDebitSchedule(ctx, shared.InvoiceCreatedEvent{ClientID: 11, InvoiceID: 1, InvoiceType: shared.InvoiceTypeB2})
	assert.Nil(suite.T(), err)

	var c int
	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection").Scan(&c)
	assert.Equal(suite.T(), 1, c)
	assert.Len(suite.T(), allPayMock.called, 0)
}

func (suite *IntegrationSuite) TestService_CreateDirectDebitSchedule_skipsWhenNoBalance() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'Client', NULL);",
		"INSERT INTO finance_client (id, client_id, sop_number, payment_method, batchnumber, court_ref) VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
		`INSERT INTO public.addresses VALUES (1, 11, '["1 Test Street"]', 'Testtown', NULL, 'TE1 1ST', NULL);`,
	)

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{}
	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn, env: &Env{AllpayEnabled: true}}

	err := s.CreateDirectDebitSchedule(ctx, shared.InvoiceCreatedEvent{ClientID: 11, InvoiceID: 1, InvoiceType: shared.InvoiceTypeB2})
	assert.Nil(suite.T(), err)

	var c int
	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection").Scan(&c)
	assert.Equal(suite.T(), 0, c)
	assert.Len(suite.T(), allPayMock.called, 0)
}
