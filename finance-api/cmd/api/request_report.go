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

	err := s.validateReportRequest(reportRequest)
	if err != nil {
		return err
	}

	if reportRequest.ReportType == shared.ReportsTypeJournal {
		goLiveDate := shared.NewDate(os.Getenv("FINANCE_HUB_LIVE_DATE"))
		if !reportRequest.TransactionDate.Before(shared.NewDate(time.Now().Format("2006-01-02"))) ||
			reportRequest.TransactionDate.Before(goLiveDate) {
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

func (s *Server) validateReportRequest(reportRequest shared.ReportRequest) error {
	validationErrors := apierror.ValidationErrors{}

	switch reportRequest.ReportType {
	case shared.ReportsTypeAccountsReceivable:
		if reportRequest.AccountsReceivableType == nil {
			validationErrors["AccountsReceivableType"] = map[string]string{
				"required": "This field AccountsReceivableType needs to be looked at required",
			}
		}
	case shared.ReportsTypeJournal:
		if reportRequest.JournalType == nil {
			validationErrors["JournalType"] = map[string]string{
				"required": "This field JournalType needs to be looked at required",
			}
		}
	case shared.ReportsTypeSchedule:
		if reportRequest.ScheduleType == nil {
			validationErrors["ScheduleType"] = map[string]string{
				"required": "This field ScheduleType needs to be looked at required",
			}
		}
	case shared.ReportsTypeDebt:
		if reportRequest.DebtType == nil {
			validationErrors["DebtType"] = map[string]string{
				"required": "This field DebtType needs to be looked at required",
			}
		}
	default:
		validationErrors["ReportType"] = map[string]string{
			"required": "This field ReportType needs to be looked at required",
		}
	}

	if reportRequest.Email == "" {
		validationErrors["Email"] = map[string]string{
			"required": "This field Email needs to be looked at required",
		}
	}

	if reportRequest.ReportType == shared.ReportsTypeSchedule {
		if reportRequest.TransactionDate == nil {
			validationErrors["Date"] = map[string]string{
				"required": "This field Date needs to be looked at required",
			}
		} else if !reportRequest.TransactionDate.Before(shared.Date{Time: time.Now().Truncate(24 * time.Hour)}) {
			validationErrors["Date"] = map[string]string{
				"date-in-the-past": "This field Date needs to be looked at date-in-the-past",
			}
		} else if reportRequest.TransactionDate.Before(shared.Date{Time: s.envs.GoLiveDate}) {
			validationErrors["Date"] = map[string]string{
				"min-go-live": "This field Date needs to be looked at min-go-live",
			}
		}
	}

	if len(validationErrors) > 0 {
		return apierror.ValidationError{Errors: validationErrors}
	}

	return nil
}
