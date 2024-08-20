package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/opg-sirius-finance-hub/finance-api/internal/service"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := telemetry.LoggerFromContext(ctx)

	var event shared.Event
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		logger.Error("Unable to parse event message")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var err error
	if event.Source == shared.EventSourceSirius && event.DetailType == shared.DetailTypeDebtPositionChanged {
		if detail, ok := event.Detail.(shared.DebtPositionChangedEvent); ok {
			err = s.Service.ReapplyCredit(ctx, int32(detail.ClientID))
		}
	} else {
		logger.Error(fmt.Sprintf("Could not match event: %s %s", event.Source, event.DetailType))
		http.Error(w, "no match", http.StatusUnprocessableEntity)
		return
	}

	if err != nil {
		var e service.BadRequest
		ok := errors.As(err, &e)
		if ok {
			http.Error(w, e.Reason, http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
