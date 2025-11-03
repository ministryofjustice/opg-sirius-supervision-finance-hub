package service

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_CreateDirectDebitSchedule_SkipsDemandedFinanceClient() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, false, false, NULL, NULL, 'Client', NULL);",
		"INSERT INTO finance_client (id, client_id, sop_number, payment_method, batchnumber, court_ref) VALUES (1, 11, '1234', 'DEMANDED', NULL, '1234567T');",
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 5000, NULL, '2024-01-01', NULL, '2024-01-01')",
	)

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{}
	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn}

	err := s.CreateDirectDebitScheduleForInvoice(ctx, 11)
	assert.Nil(suite.T(), err)

	var c int
	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection").Scan(&c)
	assert.Equal(suite.T(), 0, c)
	assert.Len(suite.T(), allPayMock.called, 0)
}

func (suite *IntegrationSuite) TestService_CreateDirectDebitScheduleForInvoice() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, false, false, NULL, NULL, 'Client', NULL);",
		"INSERT INTO finance_client (id, client_id, sop_number, payment_method, batchnumber, court_ref) VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 10000, NULL, '2024-01-01', NULL, '2024-01-01')",
		"INSERT INTO invoice VALUES (2, 11, 1, 'S2', 'S200124/24', '2024-01-01', '2025-03-31', 1000, NULL, '2024-01-01', NULL, '2024-01-01')",
	)

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{}
	dispatchMock := &mockDispatch{}

	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn, dispatch: dispatchMock}

	err := s.CreateDirectDebitScheduleForInvoice(ctx, 11)

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

func (suite *IntegrationSuite) TestService_CreateDirectDebitScheduleForInvoice_scheduleAlreadyExists() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	// Construct non-working days but ensure the calculated collection date is reproducible
	// We only need days that won't clash with the 24th of the current month
	timeNow := time.Now().UTC()
	addWorkingDaysResult := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day()+14, 0, 0, 0, 0, time.UTC)
	// Ensure addWorkingDaysResult itself is a non working day to exercise first loop in AddWorkingDays
	nonWorking := []time.Time{addWorkingDaysResult}

	allPayMock := &mockAllpay{}
	govUKMock := &mockGovUK{NonWorkingDays: nonWorking}

	// Seed client and invoice first so balance and join data exist
	seeder.SeedData(
		"INSERT INTO public.persons VALUES (11, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, false, false, NULL, NULL, 'Client', NULL);",
		"INSERT INTO finance_client (id, client_id, sop_number, payment_method, batchnumber, court_ref) VALUES (1, 11, '1234', 'DIRECT DEBIT', NULL, '1234567T');",
		"INSERT INTO invoice VALUES (1, 11, 1, 'S2', 'S200123/24', '2024-01-01', '2025-03-31', 10000, NULL, '2024-01-01', NULL, '2024-01-01')",
	)

	s := Service{store: store.New(seeder.Conn), allpay: allPayMock, govUK: govUKMock, tx: seeder.Conn}

	scheduleDate, _ := s.CalculateScheduleCollectionDate(ctx)

	seeder.SeedData(
		fmt.Sprintf("INSERT INTO pending_collection VALUES (1, 1, '%s', 5000, 'PENDING', null, '2024-01-01 00:00:00', 1)", scheduleDate.Format("2006-01-02")),
		"ALTER SEQUENCE supervision_finance.pending_collection_id_seq RESTART WITH 2",
	)

	err := s.CreateDirectDebitScheduleForInvoice(ctx, 11)
	assert.Nil(suite.T(), err)

	var c int
	_ = seeder.QueryRow(ctx, "SELECT COUNT(*) FROM pending_collection").Scan(&c)
	assert.Equal(suite.T(), 1, c)
	assert.Len(suite.T(), allPayMock.called, 0)
}
