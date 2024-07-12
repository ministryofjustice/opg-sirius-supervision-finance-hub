package service

import (
	"context"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/opg-sirius-finance-hub/finance-api/internal/testhelpers"
	"github.com/stretchr/testify/suite"
	"testing"
)

type IntegrationSuite struct {
	suite.Suite
	testDB *testhelpers.TestDatabase
	ctx    context.Context
}

func (suite *IntegrationSuite) SetupTest() {
	suite.testDB = testhelpers.InitDb()
	suite.ctx = telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test"))
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}

func (suite *IntegrationSuite) TearDownSuite() {
	suite.testDB.TearDown()
}

func (suite *IntegrationSuite) AfterTest(suiteName, testName string) {
	suite.testDB.Restore()
}
