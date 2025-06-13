package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func (suite *IntegrationSuite) TestService_UpdateRefundDecision() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (2, 1, 'findme', 'DEMANDED', 1)",
		"INSERT INTO refund VALUES (1, 2, '2019-01-27', 12300, 'PENDING', '', 99, '2025-06-04 00:00:00')",
		"INSERT INTO refund VALUES (2, 2, '2020-01-01', 32100, 'PENDING', '', 99, '2025-06-04 00:00:00')",

		"INSERT INTO bank_details VALUES (1, 1, 'Clint Client', '12345678', '11-22-33');",
		"INSERT INTO bank_details VALUES (2, 2, 'Clint Client', '12345678', '11-22-33');",
	)

	s := NewService(seeder.Conn, nil, nil, nil, nil)

	type args struct {
		clientId int32
		refundId int32
		status   shared.RefundStatus
	}
	tests := []struct {
		name              string
		args              args
		removeBankDetails bool
	}{
		{
			name: "Rejected",
			args: args{
				clientId: 1,
				refundId: 1,
				status:   shared.RefundStatusRejected,
			},
			removeBankDetails: true,
		},
		{
			name: "Approved",
			args: args{
				clientId: 1,
				refundId: 2,
				status:   shared.RefundStatusApproved,
			},
			removeBankDetails: false,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			err := s.UpdateRefundDecision(suite.ctx, tt.args.clientId, tt.args.refundId, tt.args.status)
			if err != nil {
				t.Errorf("UpdateRefundDecision() error = %v", err)
				return
			}

			var refund struct {
				status     string
				decisionAt pgtype.Date
				decisionBy int
			}
			q := seeder.QueryRow(suite.ctx, "SELECT decision, decision_at, decision_by FROM refund WHERE id = $1", tt.args.refundId)
			err = q.Scan(
				&refund.status,
				&refund.decisionAt,
				&refund.decisionBy,
			)

			if err != nil {
				t.Errorf("UpdateRefundDecision() scan fail error = %v", err)
				return
			}

			assert.Equal(t, tt.args.status.Key(), refund.status)
			assert.True(t, refund.decisionAt.Valid)
			assert.Equal(t, 10, refund.decisionBy)

			var count int
			q = seeder.QueryRow(suite.ctx, "SELECT COUNT(*) FROM bank_details where refund_id = $1", tt.args.refundId)
			_ = q.Scan(&count)

			if tt.removeBankDetails {
				assert.Equal(t, 0, count)
			} else {
				assert.Equal(t, 1, count)
			}
		})
	}
}
