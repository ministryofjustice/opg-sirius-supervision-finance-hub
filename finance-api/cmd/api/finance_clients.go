package api

import (
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"net/http"
	"strconv"
)

func (s *Server) getAccountInformation(w http.ResponseWriter, r *http.Request) {
	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	accountInfo, err := s.Service.GetAccountInformation(clientId)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(accountInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		return
	}
}
