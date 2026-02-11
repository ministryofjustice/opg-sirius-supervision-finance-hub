package api

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/notify"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/service"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type mockService struct {
	accountInfo        *shared.AccountInformation
	annualBillingInfo  shared.AnnualBillingInformation
	invoices           shared.Invoices
	feeReductions      shared.FeeReductions
	invoiceReference   *shared.InvoiceReference
	invoiceAdjustments shared.InvoiceAdjustments
	feeReduction       *shared.AddFeeReduction
	cancelFeeReduction *shared.CancelFeeReduction
	ledger             *shared.AddInvoiceAdjustmentRequest
	manualInvoice      *shared.AddManualInvoice
	adjustmentTypes    []shared.AdjustmentType
	billingHistory     []shared.BillingHistory
	refunds            shared.Refunds
	addRefund          shared.AddRefund
	pendingCollection  service.PendingCollection
	expectedIds        []int
	called             []string
	errs               map[string]error
	lastCalledParams   []interface{}
}

func (s *mockService) AddCollectedPayments(ctx context.Context, date time.Time) error {
	s.called = append(s.called, "AddCollectedPayments")
	s.lastCalledParams = []interface{}{date}
	return s.errs["AddCollectedPayments"]
}

func (s *mockService) ProcessAdhocEvent(ctx context.Context) error {
	s.called = append(s.called, "ProcessAdhocEvent")
	return s.errs["ProcessAdhocEvent"]
}

func (s *mockService) UpdatePaymentMethod(ctx context.Context, clientID int32, paymentMethod shared.PaymentMethod) error {
	s.expectedIds = []int{int(clientID)}
	s.called = append(s.called, "UpdatePaymentMethod")
	return s.errs["UpdatePaymentMethod"]
}

func (s *mockService) ReapplyCredit(ctx context.Context, clientID int32, tx *store.Tx) error {
	s.expectedIds = []int{int(clientID)}
	s.called = append(s.called, "ReapplyCredit")
	return s.errs["ReapplyCredit"]
}

func (s *mockService) GetBillingHistory(ctx context.Context, id int32) ([]shared.BillingHistory, error) {
	s.expectedIds = []int{int(id)}
	s.called = append(s.called, "GetBillingHistory")
	return s.billingHistory, s.errs["GetBillingHistory"]
}

func (s *mockService) AddManualInvoice(ctx context.Context, id int32, invoice shared.AddManualInvoice) error {
	s.expectedIds = []int{int(id)}
	s.called = append(s.called, "AddManualInvoice")
	return s.errs["AddManualInvoice"]
}

func (s *mockService) GetPermittedAdjustments(ctx context.Context, id int32) ([]shared.AdjustmentType, error) {
	s.expectedIds = []int{int(id)}
	s.called = append(s.called, "GetPermittedAdjustments")
	return s.adjustmentTypes, s.errs["GetPermittedAdjustments"]
}

func (s *mockService) UpdatePendingInvoiceAdjustment(ctx context.Context, clientId int32, adjustmentId int32, status shared.AdjustmentStatus) error {
	s.expectedIds = []int{int(adjustmentId)}
	s.called = append(s.called, "UpdatePendingInvoiceAdjustment")
	return s.errs["UpdatePendingInvoiceAdjustment"]
}

func (s *mockService) AddFeeReduction(ctx context.Context, id int32, data shared.AddFeeReduction) error {
	s.expectedIds = []int{int(id)}
	s.called = append(s.called, "AddFeeReduction")
	return s.errs["AddFeeReduction"]
}

func (s *mockService) CancelFeeReduction(ctx context.Context, id int32, cancelledFeeReduction shared.CancelFeeReduction) error {
	s.expectedIds = []int{int(id)}
	s.called = append(s.called, "CancelFeeReduction")
	return s.errs["CancelFeeReduction"]
}

func (s *mockService) GetAccountInformation(ctx context.Context, id int32) (*shared.AccountInformation, error) {
	s.expectedIds = []int{int(id)}
	s.called = append(s.called, "GetAccountInformation")
	return s.accountInfo, s.errs["GetAccountInformation"]
}

func (s *mockService) GetInvoices(ctx context.Context, id int32) (shared.Invoices, error) {
	s.expectedIds = []int{int(id)}
	s.called = append(s.called, "GetInvoices")
	return s.invoices, s.errs["GetInvoices"]
}

func (s *mockService) GetFeeReductions(ctx context.Context, id int32) (shared.FeeReductions, error) {
	s.expectedIds = []int{int(id)}
	s.called = append(s.called, "GetFeeReductions")
	return s.feeReductions, s.errs["GetFeeReductions"]
}

func (s *mockService) GetInvoiceAdjustments(ctx context.Context, id int32) (shared.InvoiceAdjustments, error) {
	s.expectedIds = []int{int(id)}
	s.called = append(s.called, "GetInvoiceAdjustments")
	return s.invoiceAdjustments, s.errs["GetInvoiceAdjustments"]
}

func (s *mockService) GetRefunds(ctx context.Context, id int32) (shared.Refunds, error) {
	s.expectedIds = []int{int(id)}
	s.called = append(s.called, "GetRefunds")
	return s.refunds, s.errs["GetRefunds"]
}

func (s *mockService) GetAnnualBillingInfo(ctx context.Context) (shared.AnnualBillingInformation, error) {
	s.called = append(s.called, "GetAnnualBillingInfo")
	return s.annualBillingInfo, s.errs["GetAnnualBillingInfo"]
}

func (s *mockService) AddRefund(ctx context.Context, id int32, refund shared.AddRefund) error {
	s.expectedIds = []int{int(id)}
	s.called = append(s.called, "AddRefund")
	return s.errs["AddRefund"]
}

func (s *mockService) UpdateRefundDecision(ctx context.Context, clientId int32, refundId int32, status shared.RefundStatus) error {
	s.expectedIds = []int{int(refundId)}
	s.called = append(s.called, "UpdateRefundDecision")
	return s.errs["UpdateRefundDecision"]
}

func (s *mockService) AddInvoiceAdjustment(ctx context.Context, clientId int32, invoiceId int32, ledgerEntry *shared.AddInvoiceAdjustmentRequest) (*shared.InvoiceReference, error) {
	s.ledger = ledgerEntry
	s.expectedIds = []int{int(clientId), int(invoiceId)}
	s.called = append(s.called, "addInvoiceAdjustment")
	return s.invoiceReference, s.errs["addInvoiceAdjustment"]
}

func (s *mockService) ProcessFinanceAdminUpload(ctx context.Context, detail shared.FinanceAdminUploadEvent) error {
	s.called = append(s.called, "ProcessFinanceAdminUpload")
	return s.errs["ProcessFinanceAdminUpload"]
}

func (s *mockService) UpdateClient(ctx context.Context, clientId int32, courtRef string) error {
	s.called = append(s.called, "UpdateClient")
	return s.errs["UpdateClient"]
}

func (s *mockService) GenerateAndUploadReport(ctx context.Context, reportRequest shared.ReportRequest, requestedDate time.Time) {
	s.called = append(s.called, "GenerateAndUploadReport")
}

func (s *mockService) ProcessPayments(ctx context.Context, records [][]string, uploadType shared.ReportUploadType, bankDate shared.Date, pisNumber int) (map[int]string, error) {
	s.called = append(s.called, "ProcessPayments")
	return nil, s.errs["ProcessPayments"]
}

func (s *mockService) ProcessPaymentReversals(ctx context.Context, records [][]string, uploadType shared.ReportUploadType) (map[int]string, error) {
	s.called = append(s.called, "ProcessPaymentReversals")
	return nil, s.errs["ProcessPaymentReversals"]
}

func (s *mockService) ProcessFulfilledRefunds(ctx context.Context, records [][]string, date shared.Date) (map[int]string, error) {
	s.called = append(s.called, "ProcessFulfilledRefunds")
	return nil, s.errs["ProcessFulfilledRefunds"]
}

func (s *mockService) ProcessRefundReversals(ctx context.Context, records [][]string, date shared.Date) (map[int]string, error) {
	s.called = append(s.called, "ProcessRefundReversals")
	return nil, s.errs["ProcessRefundReversals"]
}

func (s *mockService) ProcessDirectUploadReport(ctx context.Context, filename string, fileBytes io.Reader, uploadType shared.ReportUploadType) error {
	s.called = append(s.called, "ProcessDirectUploadReport")
	return s.errs["ProcessDirectUploadReport"]
}

func (s *mockService) PostReportActions(ctx context.Context, reportType shared.ReportRequest) {
	s.called = append(s.called, "PostReportActions")
}

func (s *mockService) ExpireRefunds(ctx context.Context) error {
	s.called = append(s.called, "ExpireRefunds")
	return s.errs["ExpireRefunds"]
}

func (s *mockService) CancelDirectDebitMandate(ctx context.Context, id int32, cancelMandate shared.CancelMandate) error {
	s.called = append(s.called, "CancelDirectDebitMandate")
	return s.errs["CancelDirectDebitMandate"]
}

func (s *mockService) CreateDirectDebitMandate(ctx context.Context, id int32, createMandate shared.CreateMandate) error {
	s.called = append(s.called, "CreateDirectDebitMandate")
	return s.errs["CreateDirectDebitMandate"]
}

func (s *mockService) CreateDirectDebitSchedule(ctx context.Context, clientID int32, data shared.CreateSchedule) (service.PendingCollection, error) {
	s.called = append(s.called, "CreateDirectDebitSchedule")
	return s.pendingCollection, s.errs["CreateDirectDebitSchedule"]
}

func (s *mockService) SendDirectDebitCollectionEvent(ctx context.Context, id int32, pendingCollection service.PendingCollection) error {
	s.called = append(s.called, "SendDirectDebitCollectionEvent")
	return s.errs["SendDirectDebitCollectionEvent"]
}

func (s *mockService) CreateDirectDebitScheduleForInvoice(ctx context.Context, clientID int32) error {
	s.called = append(s.called, "CreateDirectDebitScheduleForInvoice")
	return s.errs["CreateDirectDebitScheduleForInvoice"]
}

func (s *mockService) ProcessFailedDirectDebitCollections(ctx context.Context, date time.Time) error {
	s.called = append(s.called, "ProcessFailedDirectDebitCollections")
	s.lastCalledParams = []interface{}{date}
	return s.errs["ProcessFailedDirectDebitCollections"]
}

func (s *mockService) CheckPaymentMethod(ctx context.Context, clientID int32) error {
	s.called = append(s.called, "CheckPaymentMethod")
	s.lastCalledParams = []interface{}{clientID}
	return s.errs["CheckPaymentMethod"]
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
