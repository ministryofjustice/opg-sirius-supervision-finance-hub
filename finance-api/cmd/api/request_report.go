package api

import (
	"context"
	"encoding/json"
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

	if err := json.NewDecoder(r.Body).Decode(&reportRequest); err != nil {
		return err
	}

	err := s.validateReportRequest(reportRequest)
	if err != nil {
		return err
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

func (s *Server) validateReportRequest(reportRequest shared.ReportRequest) error {
	if reportRequest.Email == "" {
		return apierror.ValidationError{
			Errors: apierror.ValidationErrors{
				"Email": {
					"required": "This field Email needs to be looked at required",
				},
			},
		}
	}

	if reportRequest.ReportType == shared.ReportsTypeSchedule {
		if reportRequest.TransactionDate == nil {
			return apierror.ValidationError{
				Errors: apierror.ValidationErrors{
					"Date": {
						"required": "This field Date needs to be looked at required",
					},
				},
			}
		} else if !reportRequest.TransactionDate.Before(shared.Date{Time: time.Now().Truncate(24 * time.Hour)}) {
			return apierror.ValidationError{
				Errors: apierror.ValidationErrors{
					"Date": {
						"date-in-the-past": "This field Date needs to be looked at date-in-the-past",
					},
				},
			}
		} else if reportRequest.TransactionDate.Before(shared.Date{Time: s.envs.GoLiveDate}) {
			return apierror.ValidationError{
				Errors: apierror.ValidationErrors{
					"Date": {
						"min-go-live": "This field Date needs to be looked at min-go-live",
					},
				},
			}
		}
	}
	return nil
}
