package seed

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/event"
	"io"
)

type dispatchStub struct{}

func (m *dispatchStub) CreditOnAccount(context.Context, event.CreditOnAccount) error {
	return nil
}

func (m *dispatchStub) FinanceAdminUploadProcessed(context.Context, event.FinanceAdminUploadProcessed) error {
	return nil
}

type fileStorageStub struct{}

func (m *fileStorageStub) GetFile(context.Context, string, string) (io.ReadCloser, error) {
	return nil, nil
}
