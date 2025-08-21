package api

import (
	"bytes"
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"io"
	"time"
)

type mockService struct {
	accountInfo        *shared.AccountInformation
	invoices           shared.Invoices
	feeReductions      shared.FeeReductions
	invoiceReference   *shared.InvoiceReference
	invoiceAdjustments shared.InvoiceAdjustments
	feeReduction       *shared.AddFeeReduction
	cancelFeeReduction *shared.CancelFeeReduction
	ledger             *shared.AddInvoiceAdjustmentRequest
	manualInvoice      *shared.AddManualInvoice
	outstandingBalance int32
	adjustmentTypes    []shared.AdjustmentType
	billingHistory     []shared.BillingHistory
	refunds            shared.Refunds
	addRefund          shared.AddRefund
	expectedIds        []int
	lastCalled         string
	err                error
}

func (s *mockService) ProcessAdhocEvent(ctx context.Context) error {
	s.lastCalled = "ProcessAdhocEvent"
	return s.err
}

func (s *mockService) UpdatePaymentMethod(ctx context.Context, clientID int32, paymentMethod shared.PaymentMethod) error {
	s.expectedIds = []int{int(clientID)}
	s.lastCalled = "UpdatePaymentMethod"
	return s.err
}

func (s *mockService) ReapplyCredit(ctx context.Context, clientID int32, tx *store.Tx) error {
	s.expectedIds = []int{int(clientID)}
	s.lastCalled = "ReapplyCredit"
	return s.err
}

func (s *mockService) GetBillingHistory(ctx context.Context, id int32) ([]shared.BillingHistory, error) {
	s.expectedIds = []int{int(id)}
	s.lastCalled = "GetBillingHistory"
	return s.billingHistory, s.err
}

func (s *mockService) AddManualInvoice(ctx context.Context, id int32, invoice shared.AddManualInvoice) error {
	s.expectedIds = []int{int(id)}
	s.lastCalled = "AddManualInvoice"
	return s.err
}

func (s *mockService) GetPermittedAdjustments(ctx context.Context, id int32) ([]shared.AdjustmentType, error) {
	s.expectedIds = []int{int(id)}
	s.lastCalled = "GetPermittedAdjustments"
	return s.adjustmentTypes, s.err
}

func (s *mockService) UpdatePendingInvoiceAdjustment(ctx context.Context, clientId int32, adjustmentId int32, status shared.AdjustmentStatus) error {
	s.expectedIds = []int{int(adjustmentId)}
	s.lastCalled = "UpdatePendingInvoiceAdjustment"
	return s.err
}

func (s *mockService) AddFeeReduction(ctx context.Context, id int32, data shared.AddFeeReduction) error {
	s.expectedIds = []int{int(id)}
	s.lastCalled = "AddFeeReduction"
	return s.err
}

func (s *mockService) CancelFeeReduction(ctx context.Context, id int32, cancelledFeeReduction shared.CancelFeeReduction) error {
	s.expectedIds = []int{int(id)}
	s.lastCalled = "CancelFeeReduction"
	return s.err
}

func (s *mockService) GetAccountInformation(ctx context.Context, id int32) (*shared.AccountInformation, error) {
	s.expectedIds = []int{int(id)}
	s.lastCalled = "GetAccountInformation"
	return s.accountInfo, s.err
}

func (s *mockService) GetPendingOutstandingBalance(ctx context.Context, id int32) (int32, error) {
	s.expectedIds = []int{int(id)}
	s.lastCalled = "GetPendingOutstandingBalance"
	return s.outstandingBalance, s.err
}

func (s *mockService) GetInvoices(ctx context.Context, id int32) (shared.Invoices, error) {
	s.expectedIds = []int{int(id)}
	s.lastCalled = "GetInvoices"
	return s.invoices, s.err
}

func (s *mockService) GetFeeReductions(ctx context.Context, id int32) (shared.FeeReductions, error) {
	s.expectedIds = []int{int(id)}
	s.lastCalled = "GetFeeReductions"
	return s.feeReductions, s.err
}

func (s *mockService) GetInvoiceAdjustments(ctx context.Context, id int32) (shared.InvoiceAdjustments, error) {
	s.expectedIds = []int{int(id)}
	s.lastCalled = "GetInvoiceAdjustments"
	return s.invoiceAdjustments, s.err
}

func (s *mockService) GetRefunds(ctx context.Context, id int32) (shared.Refunds, error) {
	s.expectedIds = []int{int(id)}
	s.lastCalled = "GetRefunds"
	return s.refunds, s.err
}

func (s *mockService) AddRefund(ctx context.Context, id int32, refund shared.AddRefund) error {
	s.expectedIds = []int{int(id)}
	s.lastCalled = "AddRefund"
	return s.err
}

func (s *mockService) UpdateRefundDecision(ctx context.Context, clientId int32, refundId int32, status shared.RefundStatus) error {
	s.expectedIds = []int{int(refundId)}
	s.lastCalled = "UpdateRefundDecision"
	return s.err
}

func (s *mockService) AddInvoiceAdjustment(ctx context.Context, clientId int32, invoiceId int32, ledgerEntry *shared.AddInvoiceAdjustmentRequest) (*shared.InvoiceReference, error) {
	s.ledger = ledgerEntry
	s.expectedIds = []int{int(clientId), int(invoiceId)}
	s.lastCalled = "addInvoiceAdjustment"
	return s.invoiceReference, s.err
}

func (s *mockService) ProcessFinanceAdminUpload(ctx context.Context, detail shared.FinanceAdminUploadEvent) error {
	s.lastCalled = "ProcessFinanceAdminUpload"
	return s.err
}

func (s *mockService) UpdateClient(ctx context.Context, clientId int32, courtRef string) error {
	s.lastCalled = "UpdateClient"
	return s.err
}

func (s *mockService) GenerateAndUploadReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) {
	s.lastCalled = "GenerateAndUploadReport"
}

func (s *mockService) ProcessPayments(ctx context.Context, records [][]string, uploadType shared.ReportUploadType, bankDate shared.Date, pisNumber int) (map[int]string, error) {
	s.lastCalled = "ProcessPayments"
	return nil, s.err
}

func (s *mockService) ProcessPaymentReversals(ctx context.Context, records [][]string, uploadType shared.ReportUploadType) (map[int]string, error) {
	s.lastCalled = "ProcessPaymentReversals"
	return nil, s.err
}

func (s *mockService) ProcessFulfilledRefunds(ctx context.Context, records [][]string, date shared.Date) (map[int]string, error) {
	s.lastCalled = "ProcessFulfilledRefunds"
	return nil, s.err
}

func (s *mockService) ProcessDirectUploadReport(ctx context.Context, filename string, fileBytes io.Reader, uploadType shared.ReportUploadType) error {
	s.lastCalled = "ProcessDirectUploadReport"
	return s.err
}

func (s *mockService) PostReportActions(ctx context.Context, reportType shared.ReportRequest) {
	s.lastCalled = "PostReportActions"
}

func (s *mockService) ExpireRefunds(ctx context.Context) error {
	s.lastCalled = "ExpireRefunds"
	return s.err
}

func (s *mockService) AddPendingCollection(ctx context.Context, clientId int32, data shared.PendingCollection) error {
	s.lastCalled = "AddPendingCollection"
	return s.err
}

type mockFileStorage struct {
	versionId  string
	bucketname string
	filename   string
	data       io.Reader
	err        error
	exists     bool
}

func (m *mockFileStorage) GetFile(ctx context.Context, bucketName string, fileName string) (io.ReadCloser, error) {
	return io.NopCloser(m.data), m.err
}

func (m *mockFileStorage) GetFileWithVersion(ctx context.Context, bucketName string, fileName string, versionID string) (io.ReadCloser, error) {
	return io.NopCloser(m.data), m.err
}

func (m *mockFileStorage) PutFile(ctx context.Context, bucketName string, fileName string, data *bytes.Buffer) (*string, error) {
	m.bucketname = bucketName
	m.filename = fileName
	m.data = data

	return &m.versionId, nil
}

// add a FileExists method to the mockFileStorage struct
func (m *mockFileStorage) FileExists(ctx context.Context, bucketName string, filename string) bool {
	return m.exists
}

func (m *mockFileStorage) FileExistsWithVersion(ctx context.Context, bucketName string, filename string, versionID string) bool {
	return m.exists
}

type mockNotify struct {
	payload notify.Payload
	err     error
}

func (n *mockNotify) Send(ctx context.Context, payload notify.Payload) error {
	n.payload = payload
	return n.err
}

type MockReports struct {
	requestedReport *shared.ReportRequest
	requestedDate   time.Time
}

func (m *MockReports) GenerateAndUploadReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) {
	m.requestedReport = &reportRequest
	m.requestedDate = requestedDate
}
