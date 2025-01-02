package service

import (
	"context"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/testhelpers"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"testing"
)

type IntegrationSuite struct {
	suite.Suite
	testDB *testhelpers.ContainerManager
	ctx    context.Context
}

func (suite *IntegrationSuite) SetupSuite() {
	suite.ctx = telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test"))
	suite.testDB = testhelpers.Init(suite.ctx, "supervision_finance")
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}

func (suite *IntegrationSuite) TearDownSuite() {
	suite.testDB.TearDown(suite.ctx)
}

func (suite *IntegrationSuite) AfterTest(suiteName, testName string) {
	suite.testDB.Restore(suite.ctx)
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

type mockFileStorage struct {
	file io.ReadCloser
	err  error
}

func (m *mockFileStorage) GetFile(ctx context.Context, bucketName string, fileName string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.file, nil
}

func (m *mockFileStorage) PutFile(ctx context.Context, bucketName string, fileName string, file io.Reader) (*string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, nil
}
