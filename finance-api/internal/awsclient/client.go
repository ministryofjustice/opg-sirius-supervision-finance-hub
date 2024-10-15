package awsclient

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"os"
)

type Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

func NewClient(ctx context.Context) (Client, error) {
	awsRegion := os.Getenv("AWS_REGION")

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(awsRegion),
	)
	if err != nil {
		return nil, err
	}

	if iamRole, ok := os.LookupEnv("AWS_IAM_ROLE"); ok {
		client := sts.NewFromConfig(cfg)
		cfg.Credentials = stscreds.NewAssumeRoleProvider(client, iamRole)
	}

	client := s3.NewFromConfig(cfg, func(u *s3.Options) {
		u.UsePathStyle = true
		u.Region = awsRegion

		endpoint := os.Getenv("AWS_S3_ENDPOINT")
		if endpoint != "" {
			u.BaseEndpoint = &endpoint
		}
	})

	return client, nil
}
