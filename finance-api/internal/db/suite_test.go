package db

import (
	"context"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/testhelpers"
	"github.com/stretchr/testify/suite"
	"testing"
)

type IntegrationSuite struct {
	suite.Suite
	cm     *testhelpers.ContainerManager
	seeder *testhelpers.Seeder
	ctx    context.Context
}

func (suite *IntegrationSuite) SetupSuite() {
	suite.ctx = telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test"))
	cm := testhelpers.Init(suite.ctx)
	suite.seeder = cm.Seeder(suite.ctx)
	suite.cm = cm
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
