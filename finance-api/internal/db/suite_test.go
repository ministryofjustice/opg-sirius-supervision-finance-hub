package db

import (
	"context"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/db/testhelpers"
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
	suite.cm = testhelpers.NewContainerManager(suite.ctx)
}

func (suite *IntegrationSuite) SetupTest() {
	suite.seeder = testhelpers.NewSeeder(suite.ctx)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}

func (suite *IntegrationSuite) TearDownSuite() {
	suite.cm.TearDown(suite.ctx)
}

func (suite *IntegrationSuite) AfterTest(suiteName, testName string) {
	err := suite.cm.Restore(suite.ctx)
	if err != nil {
		suite.T().Errorf("Failed to restore snapshot after test %s", testName)
	}
}
