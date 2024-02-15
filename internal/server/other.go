package server

import (
	"math/rand"
	"net/http"
	"strconv"
)

type OtherTab struct {
	HoldingString string
	AppVars
}

type OtherHandler struct {
	route
}

func (h OtherHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	var data OtherTab
	data.AppVars = v
	data.HoldingString = "I am a random number: " + strconv.Itoa(rand.Int())

	h.Data = data
	return h.execute(w, r)
}
