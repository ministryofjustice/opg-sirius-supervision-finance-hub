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
	"os"
	"time"
)

func (s *Server) requestReport(w http.ResponseWriter, r *http.Request) error {
	var reportRequest shared.ReportRequest
	defer r.Body.Close()

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

	if reportRequest.ReportType == shared.ReportsTypeJournal.Key() {
		goLiveDate := shared.NewDate(os.Getenv("FINANCE_HUB_LIVE_DATE"))
		if !reportRequest.DateOfTransaction.Before(shared.NewDate(time.Now().Format("2006-01-02"))) ||
			reportRequest.DateOfTransaction.Before(goLiveDate) {
			return apierror.ValidationError{Errors: apierror.ValidationErrors{
				"Date": {
					"Date": fmt.Sprintf("Date must be before today and after %s", os.Getenv("FINANCE_HUB_LIVE_DATE")),
				},
			},
			}
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
