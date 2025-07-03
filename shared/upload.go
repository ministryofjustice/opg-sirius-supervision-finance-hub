package shared

type Upload struct {
	UploadType   ReportUploadType `json:"uploadType"`
	EmailAddress string           `json:"emailAddress"`
	Base64Data   string           `json:"data"`
	Filename     string           `json:"fileName"`
	UploadDate   Date             `json:"uploadDate"`
	PisNumber    int              `json:"pisNumber"`
}
