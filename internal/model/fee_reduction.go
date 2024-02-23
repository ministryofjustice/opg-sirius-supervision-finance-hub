package model

type FeeReduction struct {
	Id           int    `json:"id"`
	Type         string `json:"type"`
	StartDate    Date   `json:"startDate"`
	EndDate      Date   `json:"endDate"`
	DateReceived Date   `json:"dateReceived"`
	Notes        string `json:"notes"`
	Deleted      bool   `json:"deleted"`
	Status       string // calculated on server
}

type FeeReductions []FeeReduction
