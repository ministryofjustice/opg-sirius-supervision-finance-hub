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

func (h *OtherHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	var data OtherTab
	data.AppVars = v

	// example of how to change the data based on how it is being fetched
	if isHxRequest(r) {
		data.HoldingString = "I am a dynamic random number fetched without reloading the page: " + strconv.Itoa(rand.Int())
	} else {
		data.HoldingString = "I have been rendered statically on initial page load"
	}

	h.Data = data
	return h.execute(w, r)
}
