package service

import (
	"context"
	"io"
)

func (s *Service) ProcessDirectUploadReport(ctx context.Context, filename string, fileBytes io.Reader) error {
	if seeker, ok := fileBytes.(io.Seeker); ok {
		_, err := seeker.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
	}

	_, err := s.fileStorage.StreamFile(ctx, s.env.AsyncBucket, filename, fileBytes)
	if err != nil {
		return err
	}
	return nil
}
