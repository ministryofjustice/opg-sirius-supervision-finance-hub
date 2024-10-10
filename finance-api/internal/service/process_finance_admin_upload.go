package service

import (
	"context"
	"encoding/csv"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/awsclient"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"strconv"
	"strings"
	"time"
)

func (s *Service) ProcessFinanceAdminUpload(ctx context.Context, bucketName string, key string) error {
	_ = s.SendEmailToNotify(ctx, "joseph.smith@digital.justice.gov.uk", "5038b3fa-9509-46a7-952b-8431e1faad16")

	client, _ := awsclient.NewClient(ctx)

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
			err := s.processMotoCardPaymentsUploadLine(ctx, record)
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

func (s *Service) processMotoCardPaymentsUploadLine(ctx context.Context, record []string) error {
	courtReference := strings.SplitN(record[0], "-", -1)[0]
	amount := parseAmount(record[2])
	parsedDate, err := time.Parse("2006-01-02 15:04:05", record[1])
	if err != nil {
		return err
	}

	if amount == 0 {
		return nil
	}

	// check if payment has already been made
	ledgerId, _ := s.store.GetLedgerForPayment(ctx, store.GetLedgerForPaymentParams{
		Caserecnumber: pgtype.Text{String: courtReference, Valid: true},
		Amount:        int32(amount),
		Type:          "MOTO card payment",
		Datetime:      pgtype.Timestamp{Time: parsedDate, Valid: true},
	})

	if ledgerId != 0 {
		return nil
	}

	ledgerId, err = s.store.CreateLedgerForCaseRecNumber(ctx, store.CreateLedgerForCaseRecNumberParams{
		Caserecnumber: pgtype.Text{String: courtReference, Valid: true},
		Amount:        int32(amount),
		Type:          "MOTO card payment",
		Status:        "APPROVED",
		CreatedBy:     pgtype.Int4{Int32: 1, Valid: true},
		Datetime:      pgtype.Timestamp{Time: parsedDate, Valid: true},
	})

	if err != nil {
		return err
	}

	invoices, err := s.store.GetInvoicesForCaseRecNumber(ctx, pgtype.Text{String: courtReference, Valid: true})
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		if !(invoice.Amount == invoice.Received) && amount > 0 {
			allocationAmount := invoice.Amount - invoice.Received
			if allocationAmount > int32(amount) {
				allocationAmount = int32(amount)
			}

			err = s.store.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
				InvoiceID: pgtype.Int4{Int32: invoice.ID, Valid: true},
				Amount:    allocationAmount,
				Status:    "ALLOCATED",
				LedgerID:  pgtype.Int4{Int32: ledgerId, Valid: true},
			})
			if err != nil {
				return err
			}

			amount -= int(allocationAmount)
		}
	}

	if amount > 0 {
		err = s.store.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			Amount:   int32(amount),
			Status:   "UNAPPLIED",
			LedgerID: pgtype.Int4{Int32: ledgerId, Valid: true},
		})

		if err != nil {
			return err
		}
	}

	return nil
}
