package server

import "github.com/opg-sirius-finance-hub/internal/model"

type HeaderData struct {
	MyDetails model.Assignee
	Client    ClientVars
}

type ClientVars struct {
	FirstName   string
	Surname     string
	Outstanding string
}
