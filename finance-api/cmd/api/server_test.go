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
	ledger             *shared.CreateLedgerEntryRequest
	manualInvoice      *shared.AddManualInvoice
	adjustmentTypes    []shared.AdjustmentType
	expectedIds        []int
	err                error
	billingHistory     []shared.BillingHistory
}

func (s *mockService) GetBillingHistory(ctx context.Context, id int) ([]shared.BillingHistory, error) {
	s.expectedIds = []int{id}
	return s.billingHistory, s.err
}

func (s *mockService) AddManualInvoice(ctx context.Context, id int, invoice shared.AddManualInvoice) error {
	s.expectedIds = []int{id}
	return s.err
}

func (s *mockService) GetPermittedAdjustments(ctx context.Context, invoiceId int) ([]shared.AdjustmentType, error) {
	s.expectedIds = []int{invoiceId}
	return s.adjustmentTypes, s.err
}

func (s *mockService) UpdatePendingInvoiceAdjustment(ctx context.Context, id int, status string) error {
	s.expectedIds = []int{id}
	return s.err
}

func (s *mockService) AddFeeReduction(ctx context.Context, id int, data shared.AddFeeReduction) error {
	s.expectedIds = []int{id}
	return s.err
}

func (s *mockService) CancelFeeReduction(ctx context.Context, id int) error {
	s.expectedIds = []int{id}
	return s.err
}

func (s *mockService) GetAccountInformation(ctx context.Context, id int) (*shared.AccountInformation, error) {
	s.expectedIds = []int{id}
	return s.accountInfo, s.err
}

func (s *mockService) GetInvoices(ctx context.Context, id int) (*shared.Invoices, error) {
	s.expectedIds = []int{id}
	return s.invoices, s.err
}

func (s *mockService) GetFeeReductions(ctx context.Context, id int) (*shared.FeeReductions, error) {
	s.expectedIds = []int{id}
	return s.feeReductions, s.err
}

func (s *mockService) GetInvoiceAdjustments(ctx context.Context, id int) (*shared.InvoiceAdjustments, error) {
	s.expectedIds = []int{id}
	return s.invoiceAdjustments, s.err
}

func (s *mockService) CreateLedgerEntry(ctx context.Context, clientId int, invoiceId int, ledgerEntry *shared.CreateLedgerEntryRequest) (*shared.InvoiceReference, error) {
	s.ledger = ledgerEntry
	s.expectedIds = []int{clientId, invoiceId}
	return s.invoiceReference, s.err
}
