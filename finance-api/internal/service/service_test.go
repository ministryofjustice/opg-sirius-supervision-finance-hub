package service

import (
	"context"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/testhelpers"
	"github.com/stretchr/testify/suite"
	"net/http"
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

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

var (
	// GetDoFunc fetches the mock client's `Do` func. Implement this within a test to modify the client's behaviour.
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

func SetUpTest() *MockClient {
	mockClient := &MockClient{}
	return mockClient
}
