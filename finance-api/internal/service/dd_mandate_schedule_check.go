package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/allpay"
)

const (
	ddMandateScheduleCheckFromDate = "2026-04-28"
	ddMandateScheduleCheckSleep    = 250 * time.Millisecond
)

func (s *Service) CheckDirectDebitMandateSchedule(ctx context.Context, logger *slog.Logger) error {
	parsedDate, err := time.Parse("2006-01-02", ddMandateScheduleCheckFromDate)
	if err != nil {
		return err
	}

	dateFrom := pgtype.Timestamp{Time: parsedDate, Valid: true}
	clients, err := s.store.GetClientsSetToDirectDebitOnOrAfterSpecifiedDate(ctx, dateFrom)
	if err != nil {
		return err
	}

	logger.Info("DD mandate & schedule check: clients found", "count", len(clients))

	var csvBuffer bytes.Buffer
	if _, err = csvBuffer.Write([]byte("\uFEFF")); err != nil {
		return fmt.Errorf("write CSV BOM: %w", err)
	}

	writer := csv.NewWriter(&csvBuffer)
	if err = writeCSVHeader(writer); err != nil {
		return err
	}

	for i, client := range clients {
		input := allpay.FetchMandateScheduleInput{
			ClientDetails: allpay.ClientDetails{
				ClientReference: client.CourtRef.String,
				Surname:         client.Surname.String,
			},
		}

		result := fetchMandateSchedule(ctx, s.allpay, input, logger)

		if !shouldWriteMissingScheduleRow(result) {
			continue
		}

		if err = writeCSVRow(writer, client.CourtRef.String, client.Surname.String, result); err != nil {
			return err
		}

		if (i+1)%50 == 0 || i+1 == len(clients) {
			logger.Info("DD mandate & schedule check: progress", "processed", i+1, "total", len(clients))
		}

		if i+1 < len(clients) {
			time.Sleep(ddMandateScheduleCheckSleep)
		}
	}

	writer.Flush()
	if err = writer.Error(); err != nil {
		return fmt.Errorf("write CSV: %w", err)
	}

	fileName := fmt.Sprintf("dd-mandate-schedule-check/%s.csv", time.Now().UTC().Format("2006-01-02T15-04-05Z"))

	versionID, err := s.fileStorage.StreamFile(ctx, s.env.AsyncBucket, fileName, io.NopCloser(bytes.NewReader(csvBuffer.Bytes())))
	if err != nil {
		return fmt.Errorf("upload CSV to S3: %w", err)
	}

	logger.Info("DD mandate & schedule check: CSV uploaded", "bucket", s.env.AsyncBucket, "key", fileName, "versionId", versionID)
	return nil
}

func fetchMandateSchedule(ctx context.Context, allpayClient AllpayClient, input allpay.FetchMandateScheduleInput, logger *slog.Logger) *allpay.MandateScheduleCheckOutput {
	result := &allpay.MandateScheduleCheckOutput{}

	mandate, err := allpayClient.FetchMandate(ctx, input)
	if err != nil {
		logger.Error("DD mandate check: unable to fetch mandate data", "courtRef", input.ClientReference, "error", err)
		result.MandateError = err.Error()
	} else {
		result.Mandate = mandate
	}

	schedule, err := allpayClient.FetchSchedule(ctx, input)
	if err != nil {
		logger.Error("DD schedule check: unable to fetch schedule data", "courtRef", input.ClientReference, "error", err)
		result.ScheduleError = err.Error()
	} else {
		result.Schedule = schedule
	}

	return result
}

func writeCSVHeader(writer *csv.Writer) error {
	return writer.Write([]string{
		"client_ref",
		"surname",
		"mandate_status",
		"mandate_error",
		"schedule_error",
	})
}

func writeCSVRow(writer *csv.Writer, courtRef string, surname string, result *allpay.MandateScheduleCheckOutput) error {
	mandateStatus := ""
	mandateError := ""
	scheduleError := ""

	if result != nil {
		mandateError = result.MandateError
		scheduleError = result.ScheduleError

		if result.Mandate != nil && len(result.Mandate.FetchMandateScheduleDataType) > 0 {
			mandateStatus = result.Mandate.FetchMandateScheduleDataType[0].Status
		}
	}

	return writer.Write([]string{
		courtRef,
		surname,
		mandateStatus,
		mandateError,
		scheduleError,
	})
}

func shouldWriteMissingScheduleRow(result *allpay.MandateScheduleCheckOutput) bool {
	if result == nil {
		return false
	}

	if result.MandateError != "" || result.ScheduleError != "" {
		return true
	}

	mandateExists := result.Mandate != nil && result.Mandate.TotalRecords > 0
	scheduleMissing := result.Schedule != nil && result.Schedule.TotalRecords == 0

	return mandateExists && scheduleMissing
}
