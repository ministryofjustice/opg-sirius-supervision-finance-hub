package service

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"io"
	"strings"
)

func (s *Service) ProcessDirectUploadReport(ctx context.Context, uploadType shared.ReportUploadType, fileBytes io.Reader) error {
	var directory string
	switch uploadType {
	case shared.ReportTypeUploadDebtChase:
		directory = "fee-chase"
	default:
		directory = ""
	}

	filename := fmt.Sprintf("%s/%s.csv", directory, strings.ToLower(uploadType.Key()))

	_, err := s.fileStorage.StreamFile(ctx, s.env.AsyncBucket, filename, io.NopCloser(fileBytes))
	if err != nil {
		return err
	}
	return nil
}
