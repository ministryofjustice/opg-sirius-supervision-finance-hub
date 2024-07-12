package service

import (
	"context"
)

func (s *Service) CancelFeeReduction(ctx context.Context, id int) error {
	_, err := s.store.CancelFeeReduction(ctx, int32(id))

	if err != nil {
		return err
	}

	return nil
}
