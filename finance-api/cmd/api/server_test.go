package api

import (
	"github.com/opg-sirius-finance-hub/shared"
)

type mockService struct {
	accountInfo        *shared.AccountInformation
	invoices           *shared.Invoices
	feeReductions      *shared.FeeReductions
	invoiceAdjustments *shared.InvoiceAdjustments
	feeReduction       *shared.AddFeeReduction
	cancelFeeReduction *shared.CancelFeeReduction
	ledger             *shared.CreateLedgerEntryRequest
	expectedIds        []int
	err                error
}

func (s *mockService) UpdatePendingInvoiceAdjustment(id int) error {
	s.expectedId = id
	return s.err
}

func (s *mockService) AddFeeReduction(id int, feeReduction shared.AddFeeReduction) error {
	s.expectedIds = []int{id}
	return s.err
}

func (s *mockService) CancelFeeReduction(id int) error {
	s.expectedIds = []int{id}
	return s.err
}

func (s *mockService) GetAccountInformation(id int) (*shared.AccountInformation, error) {
	s.expectedIds = []int{id}
	return s.accountInfo, s.err
}

func (s *mockService) GetInvoices(id int) (*shared.Invoices, error) {
	s.expectedIds = []int{id}
	return s.invoices, s.err
}

func (s *mockService) GetFeeReductions(id int) (*shared.FeeReductions, error) {
	s.expectedIds = []int{id}
	return s.feeReductions, s.err
}

func (s *mockService) GetInvoiceAdjustments(id int) (*shared.InvoiceAdjustments, error) {
	s.expectedIds = []int{id}
	return s.invoiceAdjustments, s.err
}

func (s *mockService) CreateLedgerEntry(clientId int, invoiceId int, ledger *shared.CreateLedgerEntryRequest) error {
	s.ledger = ledger
	s.expectedIds = []int{clientId, invoiceId}
	return s.err
}
