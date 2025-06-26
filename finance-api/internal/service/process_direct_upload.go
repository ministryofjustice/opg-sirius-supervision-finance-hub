package service

import (
	"bytes"
	"context"
	"io"
)

func (s *Service) ProcessDirectUploadReport(ctx context.Context, filename string, fileBytes bytes.Reader) error {
	_, err := s.fileStorage.StreamFile(ctx, s.env.AsyncBucket, filename, io.NopCloser(&fileBytes))
	if err != nil {
		return err
	}
	return nil
}
