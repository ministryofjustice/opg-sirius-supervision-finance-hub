package internal

import (
	"database/sql"
	"encoding/json"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
)

func GetFinancePerson(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client_id, _ := strconv.Atoi(r.PathValue("id"))

		rows, err := db.Query("SELECT client_id, cachedoutstandingbalance, cachedcreditbalance, payment_method FROM finance_client WHERE client_id = $1", client_id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		var f shared.HeaderAccountData
		rows.Next()

		if err := rows.Scan(&f.ClientID, &f.OutstandingBalance, &f.CreditBalance, &f.PaymentMethod); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonData, err := json.Marshal(shared.HeaderAccountData{
			ClientID:           f.ClientID,
			OutstandingBalance: f.OutstandingBalance,
			CreditBalance:      f.CreditBalance,
			PaymentMethod:      f.PaymentMethod,
		})

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
}
