package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/opg-sirius-finance-hub/apierror"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var event shared.Event
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		return apierror.BadRequestError("event", "unable to parse event", err)
	}

	if event.Source == shared.EventSourceSirius && event.DetailType == shared.DetailTypeDebtPositionChanged {
		if detail, ok := event.Detail.(shared.DebtPositionChangedEvent); ok {
			err := s.Service.ReapplyCredit(ctx, int32(detail.ClientID))
			if err != nil {
				return err
			}
		}
	} else if event.Source == shared.EventSourceSirius && event.DetailType == shared.DetailTypeClientCreated {
		if detail, ok := event.Detail.(shared.ClientCreatedEvent); ok {
			err := s.Service.UpdateClient(ctx, detail.ClientID, detail.CaseRecNumber)
			if err != nil {
				return err
			}
		}
	} else if event.Source == shared.EventSourceS3 && event.DetailType == shared.DetailTypeAWSCloudtrailEvent {
		if detail, ok := event.Detail.(shared.FinanceAdminUploadEvent); ok {
			err := s.Service.ProcessFinanceAdminUpload(ctx, detail.RequestParameters.BucketName, detail.RequestParameters.Key)
			if err != nil {
				return err
			}
		}
	} else {
		return apierror.BadRequestError("event", fmt.Sprintf("could not match event: %s %s", event.Source, event.DetailType), errors.New("no match"))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
