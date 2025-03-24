package service

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) ProcessPaymentReversalUpload(ctx context.Context, event shared.FinanceAdminUploadEvent) error {
	// extract payments from the file
	// match payments and validate:
	//   - payment exists
	//   - matched by date or PIS if cheque
	//   - if misapplied, new client exists
	// for each payment:
	//   - create ledger
	//   - create the reversed allocation
	//   - update the payment with the reversed allocation
	//   - run reapply
	//   - if misapplied:
	//		- create ledger for correct client
	//		- create allocation for correct client
	//		- run reapply
	return nil
}
