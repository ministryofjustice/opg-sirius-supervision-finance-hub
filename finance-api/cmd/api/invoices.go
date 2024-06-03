package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) getInvoices(w http.ResponseWriter, r *http.Request) {
	clientId, _ := strconv.Atoi(r.PathValue("clientId"))
	accountInfo, err := s.Service.GetInvoices(clientId)

	if err != nil {
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

func (s *Server) getInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceId, _ := strconv.Atoi(r.PathValue("invoiceId"))
	invoice, err := s.Service.GetInvoice(invoiceId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(invoice)
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
