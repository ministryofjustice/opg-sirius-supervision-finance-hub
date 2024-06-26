package service

import (
	"context"
)

func (s *Service) CancelFeeReduction(id int) error {
	ctx := context.Background()

	_, err := s.store.CancelFeeReduction(ctx, int32(id))

	if err != nil {
		return err
	}

	return nil
}
