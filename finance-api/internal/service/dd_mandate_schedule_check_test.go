package service

import (
	"encoding/csv"
	"log/slog"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_CheckDirectDebitMandateSchedule_writesOnlyMandatesWithMissingSchedules() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO public.persons (id, firstname, surname, caserecnumber, clientstatus) VALUES (910, 'Mandy', 'MissingSchedule', 'missing-schedule', 'ACTIVE');",
		"INSERT INTO public.persons (id, firstname, surname, caserecnumber, clientstatus) VALUES (920, 'Sam', 'HasSchedule', 'has-schedule', 'ACTIVE');",
		"INSERT INTO finance_client VALUES (91001, 910, 'missing-schedule', 'DIRECT DEBIT', NULL, 'missing-schedule');",
		"INSERT INTO finance_client VALUES (92001, 920, 'has-schedule', 'DIRECT DEBIT', NULL, 'has-schedule');",
		"INSERT INTO payment_method VALUES (nextval('payment_method_id_seq'), 91001, 'DIRECT DEBIT', '2026-05-01', 1);",
		"INSERT INTO payment_method VALUES (nextval('payment_method_id_seq'), 92001, 'DIRECT DEBIT', '2026-05-01', 1);",
	)

	allpayMock := mockAllpay{
		mandates: map[string]*allpay.FetchMandateScheduleOutput{
			"missing-schedule": {
				FetchMandateScheduleDataType: allpay.FetchMandateScheduleDataType{{ClientReference: "missing-schedule", LastName: "MissingSchedule", Status: "Live"}},
				TotalRecords:                 1,
			},
			"has-schedule": {
				FetchMandateScheduleDataType: allpay.FetchMandateScheduleDataType{{ClientReference: "has-schedule", LastName: "HasSchedule", Status: "Live"}},
				TotalRecords:                 1,
			},
		},
		schedules: map[string]*allpay.FetchMandateScheduleOutput{
			"missing-schedule": {TotalRecords: 0},
			"has-schedule": {
				FetchMandateScheduleDataType: allpay.FetchMandateScheduleDataType{{Amount: 10000, ClientReference: "has-schedule", LastName: "HasSchedule", ScheduleDate: "2026-07-24", Status: "Live"}},
				TotalRecords:                 1,
			},
		},
	}
	fileStorage := mockFileStorage{}

	s := &Service{
		store:       store.New(seeder.Conn),
		allpay:      &allpayMock,
		fileStorage: &fileStorage,
		tx:          seeder.Conn,
		env:         &Env{AsyncBucket: "async-bucket"},
	}

	err := s.CheckDirectDebitMandateSchedule(ctx, slog.Default())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "async-bucket", fileStorage.bucket)
	assert.True(suite.T(), strings.HasPrefix(fileStorage.key, "dd-mandate-schedule-check/"))
	assert.Equal(suite.T(), []string{"FetchMandate", "FetchSchedule", "FetchMandate", "FetchSchedule"}, allpayMock.called)

	csvBody := strings.TrimPrefix(fileStorage.body, "\ufeff")
	records, err := csv.NewReader(strings.NewReader(csvBody)).ReadAll()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), [][]string{
		{"client_ref", "surname", "mandate_status", "mandate_error", "schedule_error"},
		{"missing-schedule", "MissingSchedule", "Live", "", ""},
	}, records)
}

func TestShouldWriteMissingScheduleRow(t *testing.T) {
	tests := []struct {
		name   string
		result *allpay.MandateScheduleCheckOutput
		want   bool
	}{
		{name: "nil result", result: nil, want: false},
		{name: "mandate exists and schedule missing", result: &allpay.MandateScheduleCheckOutput{
			Mandate:  &allpay.FetchMandateScheduleOutput{TotalRecords: 1},
			Schedule: &allpay.FetchMandateScheduleOutput{TotalRecords: 0},
		}, want: true},
		{name: "mandate and schedule exist", result: &allpay.MandateScheduleCheckOutput{
			Mandate:  &allpay.FetchMandateScheduleOutput{TotalRecords: 1},
			Schedule: &allpay.FetchMandateScheduleOutput{TotalRecords: 1},
		}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shouldWriteMissingScheduleRow(tt.result))
		})
	}
}
