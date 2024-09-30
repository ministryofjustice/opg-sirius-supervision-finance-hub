package api

import (
	"context"
	"github.com/opg-sirius-finance-hub/shared"
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

func (s *mockService) ProcessFinanceAdminUpload(ctx context.Context, bucketName string, key string) error {
	s.lastCalled = "ProcessFinanceAdminUpload"
	return s.err
}
