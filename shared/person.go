package shared

type Person struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstname"`
	Surname   string `json:"surname"`
	CourtRef  string `json:"caseRecNumber"`
}
