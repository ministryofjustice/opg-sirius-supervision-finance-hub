package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"strconv"
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

func (suite *IntegrationSuite) Test_processMotoCardPaymentsUploadLine() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'invoice-1', 'DEMANDED', NULL, '1234');",
		"INSERT INTO finance_client VALUES (2, 2, 'invoice-2', 'DEMANDED', NULL, '12345');",
		"INSERT INTO invoice VALUES (1, 1, 1, 'AD', 'AD11223/19', '2023-04-01', '2025-03-31', 15000, null, '2024-03-31', 11, '2024-03-31', null, null, null, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (2, 2, 2, 'AD', 'AD11224/19', '2023-04-01', '2025-03-31', 10000, null, '2024-03-31', 11, '2024-03-31', null, null, null, '2024-03-31 00:00:00', '99');",
	)

	dispatch := &mockDispatch{}
	client := SetUpTest()
	s := NewService(client, conn.Conn, dispatch, nil)

	tests := []struct {
		name                      string
		record                    []string
		expectedClientId          int
		expectedLedgerAllocations []createdLedgerAllocation
		expectedFailedLines       map[int]string
		want                      error
	}{
		{
			name: "Underpayment",
			record: []string{
				"1234-1",
				"2024-01-17 10:15:39",
				"100",
			},
			expectedClientId: 1,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					10000,
					"MOTO card payment",
					"CONFIRMED",
					time.Date(2024, 1, 17, 10, 15, 39, 0, time.UTC),
					10000,
					"ALLOCATED",
					1,
				},
			},
			expectedFailedLines: nil,
			want:                nil,
		},
		{
			name: "Overpayment",
			record: []string{
				"12345",
				"2024-01-17 15:30:27",
				"250.1",
			},
			expectedClientId: 2,
			expectedLedgerAllocations: []createdLedgerAllocation{
				{
					25010,
					"MOTO card payment",
					"CONFIRMED",
					time.Date(2024, 1, 17, 15, 30, 27, 0, time.UTC),
					10000,
					"ALLOCATED",
					2,
				},
				{
					25010,
					"MOTO card payment",
					"CONFIRMED",
					time.Date(2024, 1, 17, 15, 30, 27, 0, time.UTC),
					15010,
					"UNAPPLIED",
					0,
				},
			},
			expectedFailedLines: nil,
			want:                nil,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var failedLines map[int]string
			err := s.processMotoCardPaymentsUploadLine(context.Background(), tt.record, 1, &failedLines)
			assert.Equal(t, tt.want, err)
			assert.Equal(t, tt.expectedFailedLines, failedLines)

			var createdLedgerAllocations []createdLedgerAllocation

			rows, _ := conn.Query(suite.ctx,
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
