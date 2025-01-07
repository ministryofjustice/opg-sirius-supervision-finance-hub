package api

import (
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
	"os"
)

func (s *Server) checkDownload(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	logger := telemetry.LoggerFromContext(ctx)
	uid := r.URL.Query().Get("uid")

	var downloadRequest shared.DownloadRequest
	err := downloadRequest.Decode(uid)
	if err != nil {
		return err
	}

	exists := s.fileStorage.FileExists(ctx, os.Getenv("REPORTS_S3_BUCKET"), downloadRequest.Key, downloadRequest.VersionId)
	if !exists {
		logger.Error("unable to locate file for download")
		return apierror.NotFound{}
	}

	return nil
}
