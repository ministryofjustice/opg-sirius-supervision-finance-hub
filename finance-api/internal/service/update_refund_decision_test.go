package service

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func (suite *IntegrationSuite) TestService_UpdateRefundDecision() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (2, 1, 'findme', 'DEMANDED', 1)",
		"INSERT INTO refund VALUES (1, 2, '2019-01-27', 12300, 'PENDING', '', 99, '2025-06-04 00:00:00')",
		"INSERT INTO refund VALUES (2, 2, '2020-01-01', 32100, 'PENDING', '', 99, '2025-06-04 00:00:00')",
		"INSERT INTO refund VALUES (3, 2, '2020-01-01', 32100, 'APPROVED', '', 99, '2025-06-04 00:00:00', 99, '2025-06-04 00:00:00', '2025-06-04 00:00:00')",

		"INSERT INTO bank_details VALUES (1, 1, 'Clint Client', '12345678', '11-22-33');",
		"INSERT INTO bank_details VALUES (2, 2, 'Clint Client', '12345678', '11-22-33');",
		"INSERT INTO bank_details VALUES (3, 3, 'Clint Client', '12345678', '11-22-33');",
	)

	s := Service{store: store.New(seeder.Conn), tx: seeder.Conn}

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
		{
			name: "Cancelled",
			args: args{
				clientId: 1,
				refundId: 3,
				status:   shared.RefundStatusCancelled,
			},
			removeBankDetails: true,
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
				status      string
				decisionAt  pgtype.Date
				decisionBy  int
				cancelledAt pgtype.Date
				cancelledBy int
			}

			if tt.args.status == shared.RefundStatusCancelled {
				q := seeder.QueryRow(suite.ctx, "SELECT cancelled_at, cancelled_by FROM refund WHERE id = $1", tt.args.refundId)
				_ = q.Scan(
					&refund.cancelledAt,
					&refund.cancelledBy,
				)
				assert.True(t, refund.cancelledAt.Valid)
				assert.Equal(t, 10, refund.cancelledBy)
			} else {
				q := seeder.QueryRow(suite.ctx, "SELECT decision, decision_at, decision_by, cancelled_at, cancelled_by FROM refund WHERE id = $1", tt.args.refundId)
				_ = q.Scan(
					&refund.status,
					&refund.decisionAt,
					&refund.decisionBy,
					&refund.cancelledAt,
					&refund.cancelledBy,
				)

				assert.Equal(t, tt.args.status.Key(), refund.status)
				assert.True(t, refund.decisionAt.Valid)
				assert.Equal(t, 10, refund.decisionBy)
			}

			var count int
			q := seeder.QueryRow(suite.ctx, "SELECT COUNT(*) FROM bank_details where refund_id = $1", tt.args.refundId)
			_ = q.Scan(&count)

			if tt.removeBankDetails {
				assert.Equal(t, 0, count)
			} else {
				assert.Equal(t, 1, count)
			}
		})
	}
}
