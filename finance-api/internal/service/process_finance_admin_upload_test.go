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
	pisNumber        int
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

	s := NewService(seeder.Conn, nil, fileStorage, notifyClient, nil)

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
					"unable to download report",
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

			err := s.ProcessFinanceAdminUpload(suite.ctx, shared.FinanceAdminUploadEvent{
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
		"INSERT INTO finance_client VALUES (3, 3, 'invoice-3', 'DEMANDED', NULL, '123456');",
		"INSERT INTO invoice VALUES (1, 1, 1, 'AD', 'AD11223/19', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (2, 2, 2, 'AD', 'AD11224/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (3, 3, 3, 'AD', 'AD11225/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (4, 3, 3, 'AD', 'AD11226/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
	)

	dispatch := &mockDispatch{}
	s := NewService(seeder.Conn, dispatch, nil, nil, nil)

	tests := []struct {
		name                      string
		records                   [][]string
		bankDate                  shared.Date
		pisNumber                 int
		expectedClientId          int
		expectedLedgerAllocations []createdLedgerAllocation
		expectedFailedLines       map[int]string
		want                      error
	}{
		{
			name: "Underpayment",
			records: [][]string{
				{"Ordercode", "BankDate", "Amount"},
				{
					"1234-1",
					"2024-01-01 10:15:39",
					"100",
				},
			},
			bankDate:         shared.NewDate("2024-01-17"),
			pisNumber:        12,
			expectedClientId: 1,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					10000,
					"MOTO CARD PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 10, 15, 39, 0, time.UTC),
					10000,
					"ALLOCATED",
					1,
					12,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "Overpayment",
			records: [][]string{
				{"Ordercode", "BankDate", "Amount"},
				{
					"12345",
					"2024-01-01 15:30:27",
					"250.1",
				},
			},
			bankDate:         shared.NewDate("2024-01-17"),
			pisNumber:        150,
			expectedClientId: 2,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					25010,
					"MOTO CARD PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 15, 30, 27, 0, time.UTC),
					10000,
					"ALLOCATED",
					2,
					150,
				},
				{
					25010,
					"MOTO CARD PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 15, 30, 27, 0, time.UTC),
					-15010,
					"UNAPPLIED",
					0,
					0,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "Underpayment with multiple invoices",
			records: [][]string{
				{"Ordercode", "BankDate", "Amount"},
				{
					"123456",
					"2024-01-01 15:30:27",
					"50",
				},
			},
			bankDate:         shared.NewDate("2024-01-17"),
			expectedClientId: 3,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					5000,
					"MOTO CARD PAYMENT",
					"CONFIRMED",
					time.Date(2024, 1, 1, 15, 30, 27, 0, time.UTC),
					5000,
					"ALLOCATED",
					3,
					0,
				},
			},
			expectedFailedLines: map[int]string{},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var failedLines map[int]string
			failedLines, err := s.processPayments(suite.ctx, tt.records, "PAYMENTS_MOTO_CARD", tt.bankDate, shared.Nillable[int]{Value: tt.pisNumber, Valid: true})
			assert.Equal(t, tt.want, err)
			assert.Equal(t, tt.expectedFailedLines, failedLines)

			var createdLedgerAllocations []createdLedgerAllocation

			rows, _ := seeder.Query(suite.ctx,
				`SELECT l.amount, l.type, l.status, l.datetime, la.amount, la.status, la.invoice_id, l.pis_number
						FROM ledger l
						LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
					WHERE l.finance_client_id = $1`, tt.expectedClientId)

			for rows.Next() {
				var r createdLedgerAllocation
				_ = rows.Scan(&r.ledgerAmount, &r.ledgerType, &r.ledgerStatus, &r.datetime, &r.allocationAmount, &r.allocationStatus, &r.invoiceId, &r.pisNumber)
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

func Test_getPaymentDetails(t *testing.T) {
	tests := []struct {
		name                   string
		record                 []string
		uploadType             string
		ledgerType             string
		index                  int
		failedLines            map[int]string
		expectedPaymentDetails shared.PaymentDetails
		expectedFailedLines    map[int]string
	}{
		{
			name:       "Moto card",
			record:     []string{"12345678", "2025-01-02 12:04:32", "320.00"},
			uploadType: "PAYMENTS_MOTO_CARD",
			ledgerType: "Payments Moto Card",
			index:      0,
			expectedPaymentDetails: shared.PaymentDetails{
				Amount:       32000,
				ReceivedDate: time.Date(2025, time.January, 2, 12, 4, 32, 0, time.UTC),
				CourtRef:     "12345678",
				LedgerType:   "Payments Moto Card",
				BankDate:     time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "BACS",
			record:     []string{"", "", "", "", "01/06/2024", "", "20.50", "", "", "", "87654321"},
			uploadType: "PAYMENTS_SUPERVISION_BACS",
			ledgerType: "Payments Supervision BACS",
			index:      0,
			expectedPaymentDetails: shared.PaymentDetails{
				Amount:       2050,
				ReceivedDate: time.Date(2024, time.June, 1, 0, 0, 0, 0, time.UTC),
				CourtRef:     "87654321",
				LedgerType:   "Payments Supervision BACS",
				BankDate:     time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "CHEQUE",
			record:     []string{"23145746", "", "541.02", "", "31/10/2024"},
			uploadType: "PAYMENTS_SUPERVISION_CHEQUE",
			ledgerType: "Payments Supervision Cheque",
			index:      0,
			expectedPaymentDetails: shared.PaymentDetails{
				Amount:       54102,
				ReceivedDate: time.Date(2024, time.October, 31, 0, 0, 0, 0, time.UTC),
				CourtRef:     "23145746",
				LedgerType:   "Payments Supervision Cheque",
				BankDate:     time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:                "Amount parse error returns failed line",
			record:              []string{"23145746", "2024-01-01 00:00:00", "five hundred pounds!!!"},
			uploadType:          "PAYMENTS_MOTO_CARD",
			index:               0,
			failedLines:         map[int]string{},
			expectedFailedLines: map[int]string{0: "AMOUNT_PARSE_ERROR"},
		},
		{
			name:                "Date time parse error returns failed line",
			record:              []string{"23145746", "yesterday", "200"},
			uploadType:          "PAYMENTS_MOTO_CARD",
			index:               0,
			failedLines:         map[int]string{},
			expectedFailedLines: map[int]string{0: "DATE_TIME_PARSE_ERROR"},
		},
		{
			name:                "Date parse error returns failed line",
			record:              []string{"23145746", "", "200", "", "yesterday"},
			uploadType:          "PAYMENTS_SUPERVISION_CHEQUE",
			index:               0,
			failedLines:         map[int]string{},
			expectedFailedLines: map[int]string{0: "DATE_PARSE_ERROR"},
		},
		{
			name:                "Failed line adds to existing failed lines",
			record:              []string{"23145746", "yesterday", "200"},
			uploadType:          "PAYMENTS_MOTO_CARD",
			index:               1,
			failedLines:         map[int]string{0: "AMOUNT_PARSE_ERROR"},
			expectedFailedLines: map[int]string{0: "AMOUNT_PARSE_ERROR", 1: "DATE_TIME_PARSE_ERROR"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentDetails := getPaymentDetails(tt.record, tt.uploadType, shared.NewDate("01/01/2025"), tt.ledgerType, tt.index, &tt.failedLines, shared.Nillable[int]{Valid: false})
			assert.Equal(t, tt.expectedPaymentDetails, paymentDetails)
			assert.Equal(t, tt.expectedFailedLines, tt.failedLines)
		})
	}
}
func Test_formatFailedLines(t *testing.T) {
	tests := []struct {
		name        string
		failedLines map[int]string
		want        []string
	}{
		{
			name:        "Empty",
			failedLines: map[int]string{},
			want:        []string(nil),
		},
		{
			name: "Unsorted lines",
			failedLines: map[int]string{
				5: "DATE_PARSE_ERROR",
				3: "CLIENT_NOT_FOUND",
				8: "DUPLICATE_PAYMENT",
				1: "DUPLICATE_PAYMENT",
			},
			want: []string{
				"Line 1: Duplicate payment line",
				"Line 3: Could not find a client with this court reference",
				"Line 5: Unable to parse date - please use the format DD/MM/YYYY",
				"Line 8: Duplicate payment line",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formattedLines := formatFailedLines(tt.failedLines)
			assert.Equal(t, tt.want, formattedLines)
		})
	}
}
