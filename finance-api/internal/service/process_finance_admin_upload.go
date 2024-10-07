package service

import (
	"context"
	"encoding/csv"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"os"
	"strconv"
	"strings"
	"time"
)

func (s *Service) ProcessFinanceAdminUpload(ctx context.Context, bucketName string, key string) error {
	awsRegion := os.Getenv("AWS_REGION")

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(awsRegion),
	)
	if err != nil {
		return err
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

	output, err := client.GetObject(ctx, &s3.GetObjectInput{
		Key:    aws.String(key),
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		return err
	}

	csvReader := csv.NewReader(output.Body)
	records, err := csvReader.ReadAll()
	if err != nil {
		return err
	}

	for index, record := range records {
		if index != 0 {
			err := s.processMotoCardPaymentsUploadLine(record)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func parseAmount(amount string) int {
	index := strings.Index(amount, ".")

	if index != -1 && len(amount)-index == 2 {
		amount = amount + "0"
	} else if index == -1 {
		amount = amount + "00"
	}

	intAmount, _ := strconv.Atoi(strings.Replace(amount, ".", "", 1))
	return intAmount
}

func (s *Service) processMotoCardPaymentsUploadLine(record []string) error {
	courtReference := strings.SplitN(record[0], "-", -1)[0]
	amount := parseAmount(record[2])
	parsedDate, err := time.Parse("2006-01-02 15:04:05", record[1])

	if err != nil {
		return err
	}

	invoices, err := s.store.GetInvoicesForCaseRecNumber(context.Background(), pgtype.Text{String: courtReference, Valid: true})

	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		if !(invoice.Amount == invoice.Received) {
			ledgerAmount := invoice.Amount - invoice.Received
			if ledgerAmount > int32(amount) {
				ledgerAmount = int32(amount)
			}

			ledger := store.CreateLedgerForCaseRecNumberParams{
				Caserecnumber: pgtype.Text{String: courtReference, Valid: true},
				Amount:        int32(amount),
				Type:          "Online card payment",
				Status:        "APPROVED",
				CreatedBy:     pgtype.Int4{Int32: 1},
				Datetime:      pgtype.Timestamp{Time: parsedDate, Valid: true},
			}

			ledgerId, err := s.store.CreateLedgerForCaseRecNumber(context.Background(), ledger)

			if err != nil {
				return err
			}

			allocation := []store.CreateLedgerAllocationParams{
				{
					InvoiceID: pgtype.Int4{Int32: invoice.ID, Valid: true},
					Amount:    ledgerAmount,
					Status:    "ALLOCATED",
					Notes:     pgtype.Text{},
					LedgerID:  pgtype.Int4{Int32: ledgerId, Valid: true},
				},
			}

			_, err = s.store.CreateLedgerAllocation(context.Background(), allocation[0])
			if err != nil {
				return err
			}

			amount -= int(ledgerAmount)
		}
	}

	// These are payments - we need to basically apply these to the client's account.
	// They should be allocated to the client's invoices, starting with the oldest's invoice (by raised date)

	return nil
}
