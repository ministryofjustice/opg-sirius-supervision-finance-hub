package api

import (
	"errors"
	"fmt"
	"github.com/aws/smithy-go"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"io"
	"net/http"
)

func (s *Server) download(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	logger := s.Logger(ctx)
	uid := r.URL.Query().Get("uid")

	var downloadRequest shared.DownloadRequest
	err := downloadRequest.Decode(uid)
	if err != nil {
		return err
	}

	var result io.ReadCloser

	if downloadRequest.Key == "Fee_Accrual.csv" {
		result, err = s.fileStorage.GetFile(ctx, s.envs.ReportsBucket, downloadRequest.Key)
	} else {
		result, err = s.fileStorage.GetFileWithVersion(ctx, s.envs.ReportsBucket, downloadRequest.Key, downloadRequest.VersionId)
	}

	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NoSuchKey" {
				return apierror.NotFoundError(err)
			}
		}
		logger.Error("failed to get object from S3", "err", err)
		return fmt.Errorf("failed to get object from S3: %w", err)
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadRequest.Key))
	w.Header().Set("Content-Type", "text/csv")

	// Stream the S3 object to the response writer using io.Copy
	_, err = io.Copy(w, result)

	return err
}
