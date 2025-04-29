package filestorage

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"io"
)

type S3Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	Options() s3.Options
}

type Uploader interface {
	Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*manager.Uploader)) (*manager.UploadOutput, error)
}

type Client struct {
	s3       S3Client
	kmsKey   string
	uploader Uploader
}

func NewClient(ctx context.Context, region string, iamRole string, endpoint string, kmsKey string) (*Client, error) {
	awsRegion := region

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(awsRegion),
	)
	if err != nil {
		return nil, err
	}

	if iamRole != "" {
		client := sts.NewFromConfig(cfg)
		cfg.Credentials = stscreds.NewAssumeRoleProvider(client, iamRole)
	}

	s3Client := s3.NewFromConfig(cfg, func(u *s3.Options) {
		u.UsePathStyle = true
		u.Region = awsRegion

		if endpoint != "" {
			u.BaseEndpoint = &endpoint
		}
	})

	uploader := manager.NewUploader(s3Client)

	return &Client{
		s3:       s3Client,
		kmsKey:   kmsKey,
		uploader: uploader,
	}, nil
}

func (c *Client) GetFile(ctx context.Context, bucketName string, filename string) (io.ReadCloser, error) {
	output, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
	})
	if err != nil {
		return nil, err
	}

	return output.Body, nil
}

func (c *Client) GetFileWithVersion(ctx context.Context, bucketName string, filename string, versionID string) (io.ReadCloser, error) {
	output, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket:    aws.String(bucketName),
		Key:       aws.String(filename),
		VersionId: aws.String(versionID),
	})
	if err != nil {
		return nil, err
	}

	return output.Body, nil
}

func (c *Client) StreamFile(ctx context.Context, bucketName string, fileName string, stream io.ReadCloser) (*string, error) {
	output, err := c.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:               aws.String(bucketName),
		Key:                  aws.String(fileName),
		Body:                 stream,
		ContentType:          aws.String("text/csv"),
		ServerSideEncryption: "aws:kms",
		SSEKMSKeyId:          aws.String(c.kmsKey),
	})

	if err != nil {
		return nil, err
	}

	return output.VersionID, err
}

func (c *Client) FileExists(ctx context.Context, bucketName string, filename string) bool {
	_, err := c.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
	})
	return err == nil
}

func (c *Client) FileExistsWithVersion(ctx context.Context, bucketName string, filename string, versionID string) bool {
	_, err := c.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:    aws.String(bucketName),
		Key:       aws.String(filename),
		VersionId: aws.String(versionID),
	})
	return err == nil
}
