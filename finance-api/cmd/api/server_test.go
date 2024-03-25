package api

import "github.com/opg-sirius-finance-hub/shared"

type mockService struct {
	accountInfo *shared.AccountInformation
	expectedId  int
	err         error
}

func (s *mockService) GetAccountInformation(id int) (*shared.AccountInformation, error) {
	s.expectedId = id
	return s.accountInfo, s.err
}
