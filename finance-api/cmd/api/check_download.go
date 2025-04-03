package api

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (s *Server) checkDownload(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	uid := r.URL.Query().Get("uid")

	var downloadRequest shared.DownloadRequest
	err := downloadRequest.Decode(uid)
	if err != nil {
		return err
	}

	var exists bool

	if downloadRequest.Key == "Fee_Accrual.csv" {
		exists = s.fileStorage.FileExists(ctx, s.envs.ReportsBucket, downloadRequest.Key)
	} else {
		exists = s.fileStorage.FileExistsWithVersion(ctx, s.envs.ReportsBucket, downloadRequest.Key, downloadRequest.VersionId)
	}

	if !exists {
		return apierror.NotFound{}
	}

	return nil
}
