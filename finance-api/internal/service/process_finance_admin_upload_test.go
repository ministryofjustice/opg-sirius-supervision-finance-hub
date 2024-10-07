package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func (suite *IntegrationSuite) Test_processMotoCardPaymentsUploadLine() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (1, 1, 'no-invoice', 'DEMANDED', NULL, '1234');",
		"INSERT INTO invoice VALUES (1, 1, 1, 'AD', 'AD11223/19', '2023-04-01', '2025-03-31', 10000, null, '2024-03-31', 11, '2024-03-31', null, null, null, '2024-03-31 00:00:00', '99');",
	)

	dispatch := &mockDispatch{}
	s := NewService(conn.Conn, dispatch)

	tests := []struct {
		name           string
		record         []string
		expectedAmount int
		expectedDate   time.Time
		want           error
	}{
		{
			name: "Ordercode with dash",
			record: []string{
				"1234-1",
				"2024-01-17 10:15:39",
				"500",
			},
			expectedAmount: 50000,
			expectedDate:   time.Date(2024, 1, 17, 10, 15, 39, 0, time.UTC),
			want:           nil,
		},
		//{
		//	name: "Ordercode with no dash",
		//	record: []string{
		//		"1234",
		//		"2024-01-17 15:30:27",
		//		"250.1",
		//	},
		//	expectedAmount: 25010,
		//	expectedDate:   time.Date(2024, 1, 17, 15, 30, 27, 0, time.UTC),
		//	want:           nil,
		//},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			err := s.processMotoCardPaymentsUploadLine(tt.record)
			assert.Equal(t, tt.want, err)

			var createdLedger struct {
				amount     int
				ledgerType string
				status     string
				datetime   time.Time
			}

			q := conn.QueryRow(suite.ctx, "SELECT amount, type, status, datetime FROM ledger ORDER BY id DESC LIMIT 1")
			_ = q.Scan(
				&createdLedger.amount,
				&createdLedger.ledgerType,
				&createdLedger.status,
				&createdLedger.datetime,
			)

			assert.Equal(t, tt.expectedAmount, createdLedger.amount)
			assert.Equal(t, tt.expectedDate, createdLedger.datetime)
			assert.Equal(t, "Online card payment", createdLedger.ledgerType)
			assert.Equal(t, "APPROVED", createdLedger.status)
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
