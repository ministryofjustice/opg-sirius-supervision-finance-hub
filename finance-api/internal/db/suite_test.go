package db

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/service"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/testhelpers"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/suite"
)

type IntegrationSuite struct {
	suite.Suite
	cm     *testhelpers.ContainerManager
	seeder *testhelpers.Seeder
	ctx    context.Context
}

type mockDispatch struct{}

func (m *mockDispatch) PaymentMethodChanged(ctx context.Context, event event.PaymentMethod) error {
	return nil
}

func (m *mockDispatch) CreditOnAccount(ctx context.Context, event event.CreditOnAccount) error {
	return nil
}

func (m *mockDispatch) RefundAdded(ctx context.Context, event event.RefundAdded) error {
	return nil
}

func (m *mockDispatch) DirectDebitScheduleFailed(ctx context.Context, event event.DirectDebitScheduleFailed) error {
	return nil
}

func (m *mockDispatch) DirectDebitCollectionFailed(ctx context.Context, event event.DirectDebitCollectionFailed) error {
	return nil
}

func (suite *IntegrationSuite) SetupSuite() {
	suite.ctx = auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
		User:    &shared.User{ID: 1},
	}
	suite.cm = testhelpers.Init(suite.ctx, "public,supervision_finance")
	seeder := suite.cm.Seeder(suite.ctx, suite.T())
	serv := service.NewService(seeder.Conn, &mockDispatch{}, nil, nil, nil, nil, nil)
	suite.seeder = seeder.WithService(serv)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}

func (suite *IntegrationSuite) TearDownSuite() {
	suite.cm.TearDown(suite.ctx)
}

func (suite *IntegrationSuite) AfterTest(suiteName, testName string) {
	suite.cm.Restore(suite.ctx)
}

func valToPtr[T any](val T) *T {
	return &val
}
