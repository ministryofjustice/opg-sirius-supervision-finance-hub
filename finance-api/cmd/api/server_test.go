package api

import "github.com/opg-sirius-finance-hub/shared"

type mockService struct {
	accountInfo *shared.AccountInformation
	invoices    *shared.Invoices
	expectedId  int
	err         error
}

func (s *mockService) GetAccountInformation(id int) (*shared.AccountInformation, error) {
	s.expectedId = id
	return s.accountInfo, s.err
}

func (s *mockService) GetInvoices(id int) (*shared.Invoices, error) {
	s.expectedId = id
	return s.invoices, s.err
}
