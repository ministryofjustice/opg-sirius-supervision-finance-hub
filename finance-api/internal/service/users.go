package service

import (
	"context"
	"github.com/opg-sirius-finance-hub/shared"
)

func (s *Service) GetCurrentUser() (*shared.Assignee, error) {
	ctx := context.Background()

	fc, err := s.Store.FinanceClient(ctx)
	if err != nil {
		return nil, err
	}

	return &shared.Assignee{
		Id: int(fc),
	}, nil
}
