package service

import (
	"context"
	"io"
)

type FileReader struct {
	io.Reader
}

func (cb *FileReader) Close() (err error) { return nil }

func (s *Service) ProcessDirectUploadReport(ctx context.Context, filename string, fileBytes io.Reader) error {
	file := FileReader{fileBytes}
	_, err := s.fileStorage.StreamFile(ctx, s.env.AsyncBucket, filename, &file)
	if err != nil {
		return err
	}
	return nil
}
