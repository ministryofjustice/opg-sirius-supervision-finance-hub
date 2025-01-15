package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"log/slog"
	"net/http"
	"time"
)

func (s *Server) requestReport(w http.ResponseWriter, r *http.Request) error {
	var reportRequest shared.ReportRequest
	defer r.Body.Close()

	fmt.Println("Server: Received report request")
	fmt.Println(r.Body)

	if err := json.NewDecoder(r.Body).Decode(&reportRequest); err != nil {
		return err
	}

	if reportRequest.Email == "" {
		return apierror.ValidationError{Errors: apierror.ValidationErrors{
			"Email": {
				"required": "This field Email needs to be looked at required",
			},
		},
		}
	}

	go func(logger *slog.Logger) {
		err := s.reports.GenerateAndUploadReport(context.Background(), reportRequest, time.Now())
		if err != nil {
			logger.Error(err.Error())
		}
	}(telemetry.LoggerFromContext(r.Context()))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	return nil
}
