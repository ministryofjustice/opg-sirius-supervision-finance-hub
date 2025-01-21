package filestorage

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockS3Client struct {
	headObjectOutput *s3.HeadObjectOutput
	headObjectError  error
	getObjectOutput  *s3.GetObjectOutput
	getObjectError   error
	putObjectOutput  *s3.PutObjectOutput
	putObjectError   error
}

func (m *mockS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	return m.headObjectOutput, m.headObjectError
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m.putObjectOutput, m.putObjectError
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
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
		name    string
		mock    *mockS3Client
		want    *s3.GetObjectOutput
		wantErr error
	}{
		{
			name: "success",
			mock: &mockS3Client{
				getObjectOutput: &s3.GetObjectOutput{},
				getObjectError:  nil,
			},
			want:    &s3.GetObjectOutput{},
			wantErr: nil,
		},
		{
			name: "fail",
			mock: &mockS3Client{
				getObjectOutput: nil,
				getObjectError:  errors.New("error"),
			},
			want:    nil,
			wantErr: fmt.Errorf("error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{s3: tt.mock}
			got, err := client.GetFileByVersion(context.Background(), "bucket", "filename", "versionID")
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestPutFile(t *testing.T) {
	versionId := "test"
	tests := []struct {
		name    string
		mock    *mockS3Client
		want    *string
		wantErr error
	}{
		{
			name: "success",
			mock: &mockS3Client{
				putObjectOutput: &s3.PutObjectOutput{VersionId: &versionId},
				putObjectError:  nil,
			},
			want:    &versionId,
			wantErr: nil,
		},
		{
			name: "fail",
			mock: &mockS3Client{
				putObjectOutput: nil,
				putObjectError:  errors.New("error"),
			},
			want:    nil,
			wantErr: fmt.Errorf("error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{s3: tt.mock}
			got, err := client.PutFile(context.Background(), "bucket", "filename", nil)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestFileExists(t *testing.T) {
	tests := []struct {
		name string
		mock *mockS3Client
		want bool
	}{
		{
			name: "success",
			mock: &mockS3Client{
				headObjectOutput: &s3.HeadObjectOutput{},
				headObjectError:  nil,
			},
			want: true,
		},
		{
			name: "fail",
			mock: &mockS3Client{
				headObjectOutput: nil,
				headObjectError:  errors.New("error"),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{s3: tt.mock}
			got := client.FileExists(context.Background(), "bucket", "filename", "versionID")
			assert.Equal(t, tt.want, got)
		})
	}
}
