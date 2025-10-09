package filestorage

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

type mockS3Client struct {
	headObjectInput  *s3.HeadObjectInput
	headObjectOutput *s3.HeadObjectOutput
	headObjectError  error
	getObjectInput   *s3.GetObjectInput
	getObjectOutput  *s3.GetObjectOutput
	getObjectError   error
}

func (m *mockS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	m.headObjectInput = params
	return m.headObjectOutput, m.headObjectError
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	m.getObjectInput = params
	return m.getObjectOutput, m.getObjectError
}

func (m *mockS3Client) Options() s3.Options {
	return s3.Options{}
}

func TestNewClient(t *testing.T) {
	got, err := NewClient(context.Background(), "eu-west-1", "role", "some-endpoint", "key")

	assert.Nil(t, err)

	assert.IsType(t, new(Client), got)
	assert.Equal(t, "eu-west-1", got.s3.Options().Region)
	assert.Equal(t, "some-endpoint", *got.s3.Options().BaseEndpoint)
	assert.Equal(t, "key", got.kmsKey)
}

func TestGetFile(t *testing.T) {
	tests := []struct {
		name           string
		mock           *mockS3Client
		bucket         string
		filename       string
		expectedInput  *s3.GetObjectInput
		expectedOutput io.ReadCloser
		expectedError  error
	}{
		{
			name:     "success",
			bucket:   "bucket-a",
			filename: "filename-b",
			mock: &mockS3Client{
				getObjectOutput: &s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader("test"))},
				getObjectError:  nil,
			},
			expectedInput: &s3.GetObjectInput{
				Bucket: aws.String("bucket-a"),
				Key:    aws.String("filename-b"),
			},
			expectedOutput: io.NopCloser(strings.NewReader("test")),
			expectedError:  nil,
		},
		{
			name: "fail",
			mock: &mockS3Client{
				getObjectOutput: nil,
				getObjectError:  errors.New("error"),
			},
			expectedInput: &s3.GetObjectInput{
				Bucket: aws.String(""),
				Key:    aws.String(""),
			},
			expectedError: fmt.Errorf("error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{s3: tt.mock}
			got, err := client.GetFile(context.Background(), tt.bucket, tt.filename)
			assert.Equal(t, tt.expectedInput, tt.mock.getObjectInput)
			assert.Equal(t, tt.expectedOutput, got)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestGetFileWithVersion(t *testing.T) {
	tests := []struct {
		name           string
		mock           *mockS3Client
		bucket         string
		filename       string
		versionId      string
		expectedInput  *s3.GetObjectInput
		expectedOutput io.ReadCloser
		expectedError  error
	}{
		{
			name:      "success",
			bucket:    "bucket-a",
			filename:  "filename-b",
			versionId: "12",
			mock: &mockS3Client{
				getObjectOutput: &s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader("test"))},
				getObjectError:  nil,
			},
			expectedInput: &s3.GetObjectInput{
				Bucket:    aws.String("bucket-a"),
				Key:       aws.String("filename-b"),
				VersionId: aws.String("12"),
			},
			expectedOutput: io.NopCloser(strings.NewReader("test")),
			expectedError:  nil,
		},
		{
			name: "fail",
			mock: &mockS3Client{
				getObjectOutput: nil,
				getObjectError:  errors.New("error"),
			},
			expectedInput: &s3.GetObjectInput{
				Bucket:    aws.String(""),
				Key:       aws.String(""),
				VersionId: aws.String(""),
			},
			expectedError: fmt.Errorf("error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{s3: tt.mock}
			got, err := client.GetFileWithVersion(context.Background(), tt.bucket, tt.filename, tt.versionId)
			assert.Equal(t, tt.expectedInput, tt.mock.getObjectInput)
			assert.Equal(t, tt.expectedOutput, got)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

type mockUploader struct {
	output *manager.UploadOutput
	err    error
}

func (m *mockUploader) Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*manager.Uploader)) (*manager.UploadOutput, error) {
	return m.output, m.err
}

func TestStreamFile(t *testing.T) {
	versionId := "test"
	tests := []struct {
		name         string
		mockUploader *mockUploader
		mockS3       *mockS3Client
		want         *string
		wantErr      error
	}{
		{
			name: "success",
			mockUploader: &mockUploader{
				output: &manager.UploadOutput{VersionID: &versionId},
			},
			want:    &versionId,
			wantErr: nil,
		},
		{
			name: "fail",
			mockUploader: &mockUploader{
				err: errors.New("error"),
			},
			want:    nil,
			wantErr: fmt.Errorf("error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{uploader: tt.mockUploader}
			got, err := client.StreamFile(context.Background(), "bucket", "filename", io.NopCloser(strings.NewReader("test")))
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestFileExists(t *testing.T) {
	tests := []struct {
		name          string
		bucket        string
		filename      string
		mock          *mockS3Client
		expectedInput *s3.HeadObjectInput
		want          bool
	}{
		{
			name:     "success",
			bucket:   "bucket-a",
			filename: "filename-b",
			mock: &mockS3Client{
				headObjectInput:  &s3.HeadObjectInput{},
				headObjectOutput: &s3.HeadObjectOutput{},
				headObjectError:  nil,
			},
			expectedInput: &s3.HeadObjectInput{
				Bucket: aws.String("bucket-a"),
				Key:    aws.String("filename-b"),
			},
			want: true,
		},
		{
			name: "fail",
			mock: &mockS3Client{
				headObjectOutput: nil,
				headObjectError:  errors.New("error"),
			},
			expectedInput: &s3.HeadObjectInput{
				Bucket: aws.String(""),
				Key:    aws.String(""),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{s3: tt.mock}
			got := client.FileExists(context.Background(), tt.bucket, tt.filename)
			assert.Equal(t, tt.expectedInput, tt.mock.headObjectInput)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFileExistsWithVersion(t *testing.T) {
	tests := []struct {
		name          string
		bucket        string
		filename      string
		versionId     string
		mock          *mockS3Client
		expectedInput *s3.HeadObjectInput
		want          bool
	}{
		{
			name:      "success",
			bucket:    "bucket-a",
			filename:  "filename-b",
			versionId: "version-c",
			mock: &mockS3Client{
				headObjectInput:  &s3.HeadObjectInput{},
				headObjectOutput: &s3.HeadObjectOutput{},
				headObjectError:  nil,
			},
			expectedInput: &s3.HeadObjectInput{
				Bucket:    aws.String("bucket-a"),
				Key:       aws.String("filename-b"),
				VersionId: aws.String("version-c"),
			},
			want: true,
		},
		{
			name: "fail",
			mock: &mockS3Client{
				headObjectOutput: nil,
				headObjectError:  errors.New("error"),
			},
			expectedInput: &s3.HeadObjectInput{
				Bucket:    aws.String(""),
				Key:       aws.String(""),
				VersionId: aws.String(""),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{s3: tt.mock}
			got := client.FileExistsWithVersion(context.Background(), tt.bucket, tt.filename, tt.versionId)
			assert.Equal(t, tt.expectedInput, tt.mock.headObjectInput)
			assert.Equal(t, tt.want, got)
		})
	}
}
