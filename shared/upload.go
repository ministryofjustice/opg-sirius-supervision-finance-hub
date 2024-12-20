package shared

import "io"

type Upload struct {
	ReportUploadType ReportUploadType `json:"reportUploadType"`
	UploadDate       Date             `json:"uploadDate"`
	Email            string           `json:"email"`
	Filename         string           `json:"filename"`
	File             []byte           `json:"file"`
}

func NewUpload(reportUploadType ReportUploadType, uploadDate string, email string, file io.Reader, filename string) (Upload, error) {
	fileTransformed, err := io.ReadAll(file)
	if err != nil {
		return Upload{}, err
	}

	upload := Upload{
		ReportUploadType: reportUploadType,
		Email:            email,
		File:             fileTransformed,
		Filename:         filename,
	}

	if uploadDate != "" {
		uploadDateFormatted := NewDate(uploadDate)
		upload.UploadDate = uploadDateFormatted
	}

	return upload, nil
}
