package api

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"net/http"
	"strconv"
)

func (s *Server) getBillingHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := telemetry.LoggerFromContext(ctx)

	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	billingHistory, err := s.Service.GetBillingHistory(ctx, clientId)

	if err != nil {
		logger.Error("getBillingHistory", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(billingHistory)
	if err != nil {
		logger.Error("getBillingHistory", "marshalling err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		return
	}
}
