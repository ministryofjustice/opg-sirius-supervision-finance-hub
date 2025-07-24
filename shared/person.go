package shared

type FeePayer struct {
	ID     string `json:"id"`
	Status string `json:"deputyStatus"`
}

type Order struct {
	Handle string `json:"handle"`
	Label  string `json:"label"`
}

type Person struct {
	ID             int       `json:"id"`
	FirstName      string    `json:"firstname"`
	Surname        string    `json:"surname"`
	CourtRef       string    `json:"caseRecNumber"`
	AddressLine1   string    `json:"addressLine1"`
	AddressLine2   string    `json:"addressLine2"`
	Town           string    `json:"town"`
	PostCode       string    `json:"postcode"`
	FeePayer       *FeePayer `json:"feePayer"`
	ActiveCaseType *Order    `json:"activeCaseType"`
}
