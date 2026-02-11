package service

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_GetAnnualBillingInformation() {
	seeder := suite.cm.Seeder(suite.ctx, nil)

	seeder.SeedData(
		"INSERT INTO supervision_finance.property (id, key, value) VALUES (1, 'AnnualBillingYear', '2025');",
		"INSERT INTO public.persons VALUES (111, NULL, NULL, 'Scheduleson', NULL, NULL, NULL, NULL, FALSE, FALSE, NULL, NULL, 'actor_client', 'ACTIVE');",
		"INSERT INTO public.cases (id, client_id, orderstatus) VALUES (101, 111, 'ACTIVE');",
		"INSERT INTO supervision_finance.finance_client (id, client_id, sop_number, payment_method, batchnumber, court_ref) VALUES (222, 111, '1234', 'DEMANDED', NULL, '1234567T');",
		"INSERT INTO invoice VALUES (111, 111, 222, 'S2', 'S200123/25', '2025-04-02', '2026-03-31', 5000, NULL, '2025-04-02', NULL, '2025-04-02')",
	)

	Store := store.New(seeder)
	tests := []struct {
		name                 string
		expectedCount        pgtype.Int8
		expectedIssuedCount  pgtype.Int8
		expectedSkippedCount pgtype.Int8
		expectedBillingYear  string
		wantErr              bool
	}{
		{
			name:                 "returns annual billing year",
			expectedCount:        pgtype.Int8{Int64: 1, Valid: true},
			expectedIssuedCount:  pgtype.Int8{Int64: 0, Valid: false},
			expectedSkippedCount: pgtype.Int8{Int64: 0, Valid: false},
			expectedBillingYear:  "2025",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Service{
				store: Store,
			}
			got, err := s.GetAnnualBillingInfo(suite.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAnnualBillingInformation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.expectedCount, got.ExpectedCount)
			assert.Equal(t, tt.expectedIssuedCount, got.IssuedCount)
			assert.Equal(t, tt.expectedSkippedCount, got.SkippedCount)
			assert.Equal(t, tt.expectedBillingYear, got.AnnualBillingYear)
		})
	}
}
