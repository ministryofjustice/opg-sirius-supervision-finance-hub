package service

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"io"
)

func (s *Service) ProcessDirectUploadReport(ctx context.Context, filename string, fileBytes io.Reader, uploadType shared.ReportUploadType) error {
	var directory string
	switch uploadType {
	case shared.ReportTypeUploadDebtChase:
		directory = "fee-chase"
	default:
		directory = ""
	}

	filePath := fmt.Sprintf("%s/%s", directory, filename)

	_, err := s.fileStorage.StreamFile(ctx, s.env.AsyncBucket, filePath, io.NopCloser(fileBytes))
	if err != nil {
		return err
	}
	return nil
}
