package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/bank-holidays.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "bank-holidays.json")
	})

	_ = http.ListenAndServe(":8080", nil)
}
