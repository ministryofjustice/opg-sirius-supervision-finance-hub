package shared

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type errorReader struct{}

func (e *errorReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestNewUploadReturnsCorrectly(t *testing.T) {
	reportUploadType := ParseReportUploadType("SomeReportType")
	email := "test@example.com"
	uploadDate := "11/05/2024"
	fileContent := []byte("file content")
	fileName := "TestFile.txt"
	fileReader := bytes.NewReader(fileContent)

	expectedUpload := Upload{
		ReportUploadType: reportUploadType,
		Email:            email,
		Filename:         fileName,
		File:             fileContent,
		UploadDate:       NewDate(uploadDate),
	}

	upload, err := NewUpload(reportUploadType, uploadDate, email, fileReader, fileName)

	assert.NoError(t, err)
	assert.Equal(t, expectedUpload, upload)
}

func TestNewUploadWithNoFileReturnsCorrectly(t *testing.T) {
	reportUploadType := ParseReportUploadType("SomeReportType")
	email := "test@example.com"
	uploadDate := ""
	fileContent := []byte("file content")
	fileName := ""
	fileReader := bytes.NewReader(fileContent)

	expectedUpload := Upload{
		ReportUploadType: reportUploadType,
		Email:            email,
		Filename:         fileName,
		File:             fileContent,
	}

	upload, err := NewUpload(reportUploadType, uploadDate, email, fileReader, fileName)

	assert.NoError(t, err)
	assert.Equal(t, expectedUpload, upload)
}

func TestNewUpload_ReturnsError(t *testing.T) {
	// Prepare input with an errorReader that always fails
	reportUploadType := ParseReportUploadType("SomeReportType")
	email := "test@example.com"
	uploadDate := "11/05/2024"
	fileName := "TestFile.txt"
	reader := &errorReader{}

	upload, err := NewUpload(reportUploadType, uploadDate, email, reader, fileName)

	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
	assert.Empty(t, upload.File)
}
