package service

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"
)

type createdLedgerAllocation struct {
	ledgerAmount     int
	ledgerType       string
	ledgerStatus     string
	datetime         time.Time
	allocationAmount int
	allocationStatus string
	invoiceId        int
}

type mockNotify struct {
	payload notify.Payload
	err     error
}

func (n *mockNotify) Send(ctx context.Context, payload notify.Payload) error {
	n.payload = payload
	return n.err
}

func (suite *IntegrationSuite) Test_processFinanceAdminUpload() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	fileStorage := &mockFileStorage{}
	fileStorage.file = io.NopCloser(strings.NewReader("test"))
	notifyClient := &mockNotify{}

	s := NewService(seeder.Conn, nil, fileStorage, notifyClient)

	tests := []struct {
		name            string
		uploadType      string
		fileStorageErr  error
		expectedPayload notify.Payload
	}{
		{
			name:       "Unknown report",
			uploadType: "test",
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingErrorTemplateId,
				Personalisation: struct {
					Error      string `json:"error"`
					UploadType string `json:"upload_type"`
				}{
					"unknown upload type",
					"",
				},
			},
		},
		{
			name:           "S3 error",
			uploadType:     "test",
			fileStorageErr: fmt.Errorf("test"),
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingErrorTemplateId,
				Personalisation: struct {
					Error      string `json:"error"`
					UploadType string `json:"upload_type"`
				}{
					"Unable to download report",
					"",
				},
			},
		},
		{
			name:       "Known report",
			uploadType: "PAYMENTS_MOTO_CARD",
			expectedPayload: notify.Payload{
				EmailAddress: "test@email.com",
				TemplateId:   notify.ProcessingSuccessTemplateId,
				Personalisation: struct {
					UploadType string `json:"upload_type"`
				}{"Payments - MOTO card"},
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			filename := "test.csv"
			emailAddress := "test@email.com"
			fileStorage.err = tt.fileStorageErr

			err := s.ProcessFinanceAdminUpload(context.Background(), shared.FinanceAdminUploadEvent{
				EmailAddress: emailAddress, Filename: filename, UploadType: tt.uploadType,
			})
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedPayload, notifyClient.payload)
		})
	}
}

func (suite *IntegrationSuite) Test_processPayments() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'invoice-1', 'DEMANDED', NULL, '1234');",
		"INSERT INTO finance_client VALUES (2, 2, 'invoice-2', 'DEMANDED', NULL, '12345');",
		"INSERT INTO invoice VALUES (1, 1, 1, 'AD', 'AD11223/19', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (2, 2, 2, 'AD', 'AD11224/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
	)

	dispatch := &mockDispatch{}
	s := NewService(seeder.Conn, dispatch, nil, nil)

	tests := []struct {
		name                      string
		records                   [][]string
		uploadedDate              shared.Date
		expectedClientId          int
		expectedLedgerAllocations []createdLedgerAllocation
		expectedFailedLines       map[int]string
		want                      error
	}{
		{
			name: "Underpayment",
			records: [][]string{
				{"Ordercode", "Date", "Amount"},
				{
					"1234-1",
					"2024-01-17 10:15:39",
					"100",
				},
			},
			uploadedDate:     shared.NewDate("2024-01-01"),
			expectedClientId: 1,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					10000,
					"MOTO CARD PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					10000,
					"ALLOCATED",
					1,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "Overpayment",
			records: [][]string{
				{"Ordercode", "Date", "Amount"},
				{
					"12345",
					"2024-01-17 15:30:27",
					"250.1",
				},
			},
			uploadedDate:     shared.NewDate("2024-01-01"),
			expectedClientId: 2,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					25010,
					"MOTO CARD PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					10000,
					"ALLOCATED",
					2,
				},
				{
					25010,
					"MOTO CARD PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					15010,
					"UNAPPLIED",
					0,
				},
			},
			expectedFailedLines: map[int]string{},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var failedLines map[int]string
			failedLines, err := s.processPayments(context.Background(), tt.records, "PAYMENTS_MOTO_CARD", tt.uploadedDate)
			assert.Equal(t, tt.want, err)
			assert.Equal(t, tt.expectedFailedLines, failedLines)

			var createdLedgerAllocations []createdLedgerAllocation

			rows, _ := seeder.Query(suite.ctx,
				"SELECT l.amount, l.type, l.status, l.datetime, la.amount, la.status, la.invoice_id "+
					"FROM ledger l "+
					"LEFT JOIN ledger_allocation la ON l.id = la.ledger_id "+
					"WHERE l.finance_client_id = $1", tt.expectedClientId)

			for rows.Next() {
				var r createdLedgerAllocation
				_ = rows.Scan(&r.ledgerAmount, &r.ledgerType, &r.ledgerStatus, &r.datetime, &r.allocationAmount, &r.allocationStatus, &r.invoiceId)
				createdLedgerAllocations = append(createdLedgerAllocations, r)
			}

			assert.Equal(t, tt.expectedLedgerAllocations, createdLedgerAllocations)
		})
	}
}

func Test_parseAmount(t *testing.T) {
	tests := []struct {
		name         string
		stringAmount string
		wantAmount   int32
		wantErr      bool
	}{
		{
			name:         "No decimal",
			stringAmount: "500",
			wantAmount:   50000,
		},
		{
			name:         "Single decimal",
			stringAmount: "500.1",
			wantAmount:   50010,
		},
		{
			name:         "Two decimals",
			stringAmount: "500.12",
			wantAmount:   50012,
		},
		{
			name:         "Unable to parse",
			stringAmount: "hehe",
			wantErr:      true,
		},
		{
			name:         "Empty string",
			stringAmount: "",
			wantAmount:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := parseAmount(tt.stringAmount)
			assert.Equal(t, tt.wantAmount, amount)

			if tt.wantErr {
				assert.IsType(t, &strconv.NumError{}, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
