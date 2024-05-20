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

const StatusPending string = "Pending"
const StatusActive string = "Active"
const StatusExpired string = "Expired"
const StatusCancelled string = "Cancelled"
