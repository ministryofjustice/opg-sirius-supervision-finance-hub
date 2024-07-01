package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/testhelpers"
	"github.com/stretchr/testify/suite"
	"log"
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

func TestName(t *testing.T) {
	i := pgtype.Int4{
		Int32: 0,
		Valid: false,
	}
	a := int(i.Int32)
	log.Print(a)
}
