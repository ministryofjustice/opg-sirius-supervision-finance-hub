package service

import (
	"context"
	"fmt"
	"log/slog"
)

func (s *Service) GetPendingOutstandingBalance(ctx context.Context, id int32) (int32, error) {
	balance, err := s.store.GetPendingOutstandingBalance(ctx, id)

	if err != nil {
		s.Logger(ctx).Error(fmt.Sprintf("Error in getting pending outstanding balance for client %d", id), slog.String("err", err.Error()))
	}

	return balance, err
}
