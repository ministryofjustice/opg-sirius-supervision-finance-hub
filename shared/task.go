package shared

type Task struct {
	ClientId int    `json:"personId"`
	Type     string `json:"type"`
	DueDate  Date   `json:"dueDate"`
	Assignee int    `json:"assigneeId"`
	Notes    string `json:"description"`
}
