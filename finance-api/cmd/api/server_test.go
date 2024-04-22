package api

import "github.com/opg-sirius-finance-hub/shared"

type mockService struct {
	accountInfo        *shared.AccountInformation
	invoices           *shared.Invoices
	feeReductions      *shared.FeeReductions
	invoiceAdjustments *shared.InvoiceAdjustments
	expectedId         int
	err                error
}

func (s *mockService) AddFeeReduction(body shared.AddFeeReduction) error {
	return s.err
}

func (s *mockService) GetAccountInformation(id int) (*shared.AccountInformation, error) {
	s.expectedId = id
	return s.accountInfo, s.err
}

func (s *mockService) GetInvoices(id int) (*shared.Invoices, error) {
	s.expectedId = id
	return s.invoices, s.err
}

func (s *mockService) GetFeeReductions(id int) (*shared.FeeReductions, error) {
	s.expectedId = id
	return s.feeReductions, s.err
}

func (s *mockService) GetInvoiceAdjustments(id int) (*shared.InvoiceAdjustments, error) {
	s.expectedId = id
	return s.invoiceAdjustments, s.err
}
