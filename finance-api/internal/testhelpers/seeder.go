package testhelpers

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Service interface {
	AddManualInvoice(ctx context.Context, clientID int32, invoice shared.AddManualInvoice) error
	AddInvoiceAdjustment(ctx context.Context, clientID int32, invoiceId int32, adjustment *shared.AddInvoiceAdjustmentRequest) (*shared.InvoiceReference, error)
	UpdatePendingInvoiceAdjustment(ctx context.Context, clientID int32, adjustmentId int32, status shared.AdjustmentStatus) error
	AddFeeReduction(ctx context.Context, clientId int32, reduction shared.AddFeeReduction) error
	ProcessPaymentsUploadLine(ctx context.Context, tx *store.Tx, details shared.PaymentDetails) (int32, error)
	ProcessReversalUploadLine(ctx context.Context, tx *store.Tx, details shared.ReversalDetails) error
	ProcessPaymentReversals(ctx context.Context, records [][]string, uploadType shared.ReportUploadType) (map[int]string, error)
	CancelFeeReduction(ctx context.Context, id int32, cancelledFeeReduction shared.CancelFeeReduction) error
	AddRefund(ctx context.Context, clientId int32, refund shared.AddRefund) error
	UpdateRefundDecision(ctx context.Context, clientId int32, refundId int32, status shared.RefundStatus) error
	PostReportActions(ctx context.Context, reportType shared.ReportRequest)
	BeginStoreTx(ctx context.Context) (*store.Tx, error)
	ProcessFulfilledRefundsLine(ctx context.Context, tx *store.Tx, refundID int32, refund shared.FulfilledRefundDetails) error
}

// Seeder contains a test database connection pool and HTTP server for API calls
type Seeder struct {
	t       *testing.T
	Conn    *pgxpool.Pool
	Service Service
}

func (s *Seeder) WithService(service Service) *Seeder {
	s.Service = service
	return s
}

func (s *Seeder) Exec(ctx context.Context, str string, i ...interface{}) (pgconn.CommandTag, error) {
	return s.Conn.Exec(ctx, str, i...)
}

func (s *Seeder) Query(ctx context.Context, str string, i ...interface{}) (pgx.Rows, error) {
	return s.Conn.Query(ctx, str, i...)
}

func (s *Seeder) QueryRow(ctx context.Context, str string, i ...interface{}) pgx.Row {
	return s.Conn.QueryRow(ctx, str, i...)
}

func (s *Seeder) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.Conn.BeginTx(ctx, pgx.TxOptions{})
}

func (s *Seeder) SeedData(data ...string) {
	ctx := context.Background()
	for _, d := range data {
		_, err := s.Exec(ctx, d)
		assert.NoError(s.t, err, "failed to seed data")
	}
}
