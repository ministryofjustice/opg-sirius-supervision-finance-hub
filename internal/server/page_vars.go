package server

import "github.com/opg-sirius-finance-hub/internal/model"

type PageVars struct {
	MyDetails model.Assignee
	Client    ClientVars
	AppVars
}

type ClientVars struct {
	FirstName   string
	Surname     string
	Outstanding string
}
