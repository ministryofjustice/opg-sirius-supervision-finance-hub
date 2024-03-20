package shared

type Person struct {
	ID        int    `json:"id"`
	Firstname string `json:"firstname"`
	Surname   string `json:"surname"`
	CourtRef  string `json:"caseRecNumber"`
}
