package shared

type FeeReduction struct {
	Id           int    `json:"id"`
	Type         string `json:"type"`
	StartDate    Date   `json:"startDate"`
	EndDate      Date   `json:"endDate"`
	DateReceived Date   `json:"dateReceived"`
	Status       string `json:"status"`
	Notes        string `json:"notes"`
}

type FeeReductions []FeeReduction
