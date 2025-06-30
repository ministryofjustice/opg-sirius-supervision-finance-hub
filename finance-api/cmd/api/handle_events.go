package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var event shared.Event
	defer unchecked(r.Body.Close)

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		return apierror.BadRequestError("event", "unable to parse event", err)
	}

	if event.Source == shared.EventSourceSirius && event.DetailType == shared.DetailTypeDebtPositionChanged {
		if detail, ok := event.Detail.(shared.DebtPositionChangedEvent); ok {
			err := s.service.ReapplyCredit(ctx, detail.ClientID)
			if err != nil {
				return err
			}
		}
	} else if event.Source == shared.EventSourceSirius && event.DetailType == shared.DetailTypeClientCreated {
		if detail, ok := event.Detail.(shared.ClientCreatedEvent); ok {
			err := s.service.UpdateClient(ctx, detail.ClientID, detail.CourtRef)
			if err != nil {
				return err
			}
		}
	} else if event.Source == shared.EventSourceFinanceAdhoc && event.DetailType == shared.DetailTypeFinanceAdhoc {
		if detail, ok := event.Detail.(shared.AdhocEvent); ok {
			err := s.processAdhocEvent(ctx, detail)
			if err != nil {
				return err
			}
		}
	} else if event.Source == shared.EventSourceInfra && event.DetailType == shared.DetailTypeScheduledEvent {
		if detail, ok := event.Detail.(shared.ScheduledEvent); ok {
			err := s.processScheduledEvent(ctx, detail)
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
