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
	AddManualInvoice(ctx context.Context, clientID int, invoice shared.AddManualInvoice) error
	AddInvoiceAdjustment(ctx context.Context, clientID int, invoiceId int, adjustment *shared.AddInvoiceAdjustmentRequest) (*shared.InvoiceReference, error)
	UpdatePendingInvoiceAdjustment(ctx context.Context, clientID int, adjustmentId int, status shared.AdjustmentStatus) error
	AddFeeReduction(ctx context.Context, clientId int, reduction shared.AddFeeReduction) error
	ProcessPaymentsUploadLine(ctx context.Context, tx *store.Tx, details shared.PaymentDetails, index int, failedLines *map[int]string) error
	BeginStoreTx(ctx context.Context) (*store.Tx, error)
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
