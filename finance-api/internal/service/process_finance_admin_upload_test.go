package service

import (
	"context"
	"github.com/stretchr/testify/assert"
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
	s := NewService(conn.Conn, dispatch)

	tests := []struct {
		name                      string
		record                    []string
		expectedClientId          int
		expectedLedgerAllocations []createdLedgerAllocation
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
					"APPROVED",
					time.Date(2024, 1, 17, 10, 15, 39, 0, time.UTC),
					10000,
					"ALLOCATED",
					1,
				},
			},
			want: nil,
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
					"APPROVED",
					time.Date(2024, 1, 17, 15, 30, 27, 0, time.UTC),
					10000,
					"ALLOCATED",
					2,
				},
				{
					25010,
					"MOTO card payment",
					"APPROVED",
					time.Date(2024, 1, 17, 15, 30, 27, 0, time.UTC),
					15010,
					"UNAPPLIED",
					0,
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			err := s.processMotoCardPaymentsUploadLine(context.Background(), tt.record)
			assert.Equal(t, tt.want, err)

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
		want         int
	}{
		{
			name:         "No decimal",
			stringAmount: "500",
			want:         50000,
		},
		{
			name:         "Single decimal",
			stringAmount: "500.1",
			want:         50010,
		},
		{
			name:         "Two decimals",
			stringAmount: "500.12",
			want:         50012,
		},
		{
			name:         "Unable to parse",
			stringAmount: "hehe",
			want:         0,
		},
		{
			name:         "Empty string",
			stringAmount: "",
			want:         0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, parseAmount(tt.stringAmount))
		})
	}
}
