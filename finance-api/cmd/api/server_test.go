package api

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"io"
	"time"
)

type mockService struct {
	accountInfo        *shared.AccountInformation
	invoices           *shared.Invoices
	feeReductions      *shared.FeeReductions
	invoiceReference   *shared.InvoiceReference
	invoiceAdjustments *shared.InvoiceAdjustments
	feeReduction       *shared.AddFeeReduction
	cancelFeeReduction *shared.CancelFeeReduction
	ledger             *shared.AddInvoiceAdjustmentRequest
	manualInvoice      *shared.AddManualInvoice
	adjustmentTypes    []shared.AdjustmentType
	billingHistory     []shared.BillingHistory
	expectedIds        []int
	lastCalled         string
	err                error
}

func (s *mockService) ReapplyCredit(ctx context.Context, clientID int32) error {
	s.expectedIds = []int{int(clientID)}
	s.lastCalled = "ReapplyCredit"
	return s.err
}

func (s *mockService) GetBillingHistory(ctx context.Context, id int) ([]shared.BillingHistory, error) {
	s.expectedIds = []int{id}
	s.lastCalled = "GetBillingHistory"
	return s.billingHistory, s.err
}

func (s *mockService) AddManualInvoice(ctx context.Context, id int, invoice shared.AddManualInvoice) error {
	s.expectedIds = []int{id}
	s.lastCalled = "AddManualInvoice"
	return s.err
}

func (s *mockService) GetPermittedAdjustments(ctx context.Context, invoiceId int) ([]shared.AdjustmentType, error) {
	s.expectedIds = []int{invoiceId}
	s.lastCalled = "GetPermittedAdjustments"
	return s.adjustmentTypes, s.err
}

func (s *mockService) UpdatePendingInvoiceAdjustment(ctx context.Context, clientId int, adjustmentId int, status shared.AdjustmentStatus) error {
	s.expectedIds = []int{adjustmentId}
	s.lastCalled = "UpdatePendingInvoiceAdjustment"
	return s.err
}

func (s *mockService) AddFeeReduction(ctx context.Context, id int, data shared.AddFeeReduction) error {
	s.expectedIds = []int{id}
	s.lastCalled = "AddFeeReduction"
	return s.err
}

func (s *mockService) CancelFeeReduction(ctx context.Context, id int, cancelledFeeReduction shared.CancelFeeReduction) error {
	s.expectedIds = []int{id}
	s.lastCalled = "CancelFeeReduction"
	return s.err
}

func (s *mockService) GetAccountInformation(ctx context.Context, id int) (*shared.AccountInformation, error) {
	s.expectedIds = []int{id}
	s.lastCalled = "GetAccountInformation"
	return s.accountInfo, s.err
}

func (s *mockService) GetInvoices(ctx context.Context, id int) (*shared.Invoices, error) {
	s.expectedIds = []int{id}
	s.lastCalled = "GetInvoices"
	return s.invoices, s.err
}

func (s *mockService) GetFeeReductions(ctx context.Context, id int) (*shared.FeeReductions, error) {
	s.expectedIds = []int{id}
	s.lastCalled = "GetFeeReductions"
	return s.feeReductions, s.err
}

func (s *mockService) GetInvoiceAdjustments(ctx context.Context, id int) (*shared.InvoiceAdjustments, error) {
	s.expectedIds = []int{id}
	s.lastCalled = "GetInvoiceAdjustments"
	return s.invoiceAdjustments, s.err
}

func (s *mockService) AddInvoiceAdjustment(ctx context.Context, clientId int, invoiceId int, ledgerEntry *shared.AddInvoiceAdjustmentRequest) (*shared.InvoiceReference, error) {
	s.ledger = ledgerEntry
	s.expectedIds = []int{clientId, invoiceId}
	s.lastCalled = "AddInvoiceAdjustment"
	return s.invoiceReference, s.err
}

func (s *mockService) ProcessFinanceAdminUpload(ctx context.Context, detail shared.FinanceAdminUploadEvent) error {
	s.lastCalled = "ProcessFinanceAdminUpload"
	return s.err
}

func (s *mockService) UpdateClient(ctx context.Context, clientId int, courtRef string) error {
	s.lastCalled = "UpdateClient"
	return s.err
}

func (s *mockService) GenerateAndUploadReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) error {
	s.lastCalled = "GenerateAndUploadReport"
	return s.err
}

type MockFileStorage struct {
	versionId      string
	bucketname     string
	filename       string
	file           io.Reader
	outgoingObject *s3.GetObjectOutput
	err            error
	exists         bool
}

func (m *MockFileStorage) GetFileByVersion(ctx context.Context, bucketName string, fileName string, versionId string) (*s3.GetObjectOutput, error) {
	return m.outgoingObject, m.err
}

func (m *MockFileStorage) PutFile(ctx context.Context, bucketName string, fileName string, file io.Reader) (*string, error) {
	m.bucketname = bucketName
	m.filename = fileName
	m.file = file

	return &m.versionId, nil
}

// add a FileExists method to the MockFileStorage struct
func (m *MockFileStorage) FileExists(ctx context.Context, bucketName string, filename string, versionID string) bool {
	return m.exists
}

type MockReports struct {
	requestedReport *shared.ReportRequest
	requestedDate   time.Time
	err             error
}

func (m *MockReports) GenerateAndUploadReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) error {
	m.requestedReport = &reportRequest
	m.requestedDate = requestedDate
	return m.err
}
