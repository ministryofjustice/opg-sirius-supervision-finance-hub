package service

import (
	"github.com/opg-sirius-finance-hub/finance-api/internal/testhelpers"
	"github.com/stretchr/testify/suite"
	"testing"
)

type IntegrationSuite struct {
	suite.Suite
	testDB *testhelpers.TestDatabase
}

func (suite *IntegrationSuite) SetupTest() {
	suite.testDB = testhelpers.InitDb()
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
