package service

import (
	"context"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
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

func (suite *IntegrationSuite) SetupSuite() {
	suite.ctx = auth.Context{
		Context: telemetry.ContextWithLogger(context.Background(), telemetry.NewLogger("finance-api-test")),
		User:    &shared.User{ID: 10},
	}
	suite.cm = testhelpers.Init(suite.ctx, "supervision_finance")
	suite.seeder = suite.cm.Seeder(suite.ctx, suite.T())
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

type mockDispatch struct {
	event any
}

func (m *mockDispatch) PaymentMethodChanged(ctx context.Context, event event.PaymentMethod) error {
	m.event = event
	return nil
}

func (m *mockDispatch) CreditOnAccount(ctx context.Context, event event.CreditOnAccount) error {
	m.event = event
	return nil
}

func (m *mockDispatch) RefundAdded(ctx context.Context, event event.RefundAdded) error {
	m.event = event
	return nil
}

func (m *mockDispatch) DirectDebitScheduleFailed(ctx context.Context, event event.DirectDebitScheduleFailed) error {
	m.event = event
	return nil
}

func (m *mockDispatch) DirectDebitCollection(ctx context.Context, event event.DirectDebitCollection) error {
	m.event = event
	return nil
}

func (m *mockDispatch) DebtChaseUploaded(ctx context.Context, event event.DebtChaseUploaded) error {
	m.event = event
	return nil
}

type mockAllpay struct {
	called           []string
	failedPayments   allpay.FailedPayments
	errs             map[string]error
	lastCalledParams []interface{}
	closureDate      time.Time
}

func (m *mockAllpay) CancelMandate(ctx context.Context, data *allpay.CancelMandateRequest) error {
	m.called = append(m.called, "CancelMandate")
	m.lastCalledParams = []interface{}{data}
	m.closureDate = data.ClosureDate
	return m.errs["CancelMandate"]
}

func (m *mockAllpay) CreateMandate(ctx context.Context, data *allpay.CreateMandateRequest) error {
	m.called = append(m.called, "CreateMandate")
	m.lastCalledParams = []interface{}{data}
	return m.errs["CreateMandate"]
}

func (m *mockAllpay) ModulusCheck(ctx context.Context, sortCode string, accountNumber string) error {
	m.called = append(m.called, "ModulusCheck")
	m.lastCalledParams = []interface{}{sortCode, accountNumber}
	return m.errs["ModulusCheck"]
}

func (m *mockAllpay) CreateSchedule(ctx context.Context, data *allpay.CreateScheduleInput) error {
	m.called = append(m.called, "CreateSchedule")
	m.lastCalledParams = []interface{}{data}
	return m.errs["CreateSchedule"]
}

func (m *mockAllpay) FetchFailedPayments(ctx context.Context, input allpay.FetchFailedPaymentsInput) (allpay.FailedPayments, error) {
	m.called = append(m.called, "FetchFailedPayments")
	m.lastCalledParams = []interface{}{input}
	return m.failedPayments, m.errs["FetchFailedPayments"]
}

func (m *mockAllpay) RemoveScheduledPayment(ctx context.Context, data *allpay.RemoveScheduledPaymentRequest) error {
	m.called = append(m.called, "RemoveScheduledPayment")
	m.lastCalledParams = []interface{}{data}
	return m.errs["RemoveScheduledPayment"]
}

type mockGovUK struct {
	called         []string
	errs           map[string]error
	NonWorkingDays []time.Time
	nWorkingDays   int
	Xday           int
	WorkingDay     time.Time
}

func (m *mockGovUK) AddWorkingDays(ctx context.Context, d time.Time, n int) (time.Time, error) {
	m.called = append(m.called, "AddWorkingDays")
	m.nWorkingDays = n
	d = time.Date(d.Year(), d.Month(), d.Day()+n, 0, 0, 0, 0, time.UTC)

	for slices.Contains(m.NonWorkingDays, d) {
		d = d.AddDate(0, 0, 1)
	}
	m.WorkingDay = d
	return d, m.errs["AddWorkingDays"]
}

func (m *mockGovUK) NextWorkingDayOnOrAfterX(ctx context.Context, d time.Time, dayOfMonth int) (time.Time, error) {
	m.called = append(m.called, "NextWorkingDayOnOrAfterX")
	m.Xday = dayOfMonth
	d = time.Date(d.Year(), d.Month(), dayOfMonth, 0, 0, 0, 0, time.UTC)

	for slices.Contains(m.NonWorkingDays, d) {
		d = d.AddDate(0, 0, 1)
	}
	m.WorkingDay = d
	return d, m.errs["NextWorkingDayOnOrAfterX"]
}
