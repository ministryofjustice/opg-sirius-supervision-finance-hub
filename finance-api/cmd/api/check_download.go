package api

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
	"os"
)

func (s *Server) checkDownload(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	uid := r.URL.Query().Get("uid")

	var downloadRequest shared.DownloadRequest
	err := downloadRequest.Decode(uid)
	if err != nil {
		return err
	}

	// TODO: Decide if this should be checked here or moved to Service layer
	exists := s.filestorage.FileExists(ctx, os.Getenv("REPORTS_S3_BUCKET"), downloadRequest.Key, downloadRequest.VersionId)
	if !exists {
		return apierror.NotFound{}
	}

	return nil
}
