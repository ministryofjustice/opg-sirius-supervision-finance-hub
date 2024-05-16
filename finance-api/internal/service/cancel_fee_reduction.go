package service

import (
	"context"
	"log"
)

func (s *Service) CancelFeeReduction(id int) error {
	ctx := context.Background()

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Println("Error rolling back cancel fee reduction transaction:", err)
		}
	}()

	transaction := s.Store.WithTx(tx)

	_, err = transaction.CancelFeeReduction(ctx, int32(id))
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
