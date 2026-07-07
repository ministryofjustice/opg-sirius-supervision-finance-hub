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

	//creates temporary CSV in memory to hold the data
	var csvBuffer bytes.Buffer
	//writes a UTF-8 BOM — Byte Order Mark — at the start of the CSV to help Excel open the CSV without character encoding issues.
	if _, err = csvBuffer.Write([]byte("\uFEFF")); err != nil {
		return fmt.Errorf("write CSV BOM: %w", err)
	}

	writer := csv.NewWriter(&csvBuffer)
	if err = writeCSVHeader(writer); err != nil {
		return err
	}

	for i, client := range clients {
		result, err := s.allpay.FetchMandateSchedule(ctx, allpay.FetchMandateScheduleInput{
			ClientDetails: allpay.ClientDetails{
				ClientReference: client.CourtRef.String,
				Surname:         client.Surname.String,
			},
		})
		if err != nil {
			logger.Error("DD mandate & schedule check: unable to fetch mandate schedule data", "courtRef", client.CourtRef.String, "error", err)
			result = &allpay.MandateScheduleCheckOutput{MandateError: err.Error()}
		}

		if !shouldWriteMissingScheduleRow(result) {
			continue
		}

		//only write CSV row for court refs that have missing schedules
		if err = writeCSVRow(writer, client.CourtRef.String, client.Surname.String, result); err != nil {
			return err
		}

		//service logs every 50 clients
		if (i+1)%50 == 0 || i+1 == len(clients) {
			logger.Info("DD mandate & schedule check: progress", "processed", i+1, "total", len(clients))
		}

		//service sleeps for 250ms between each client to avoid overwhelming the AllPay API
		if i+1 < len(clients) {
			time.Sleep(ddMandateScheduleCheckSleep)
		}
	}

	writer.Flush()
	if err = writer.Error(); err != nil {
		return fmt.Errorf("write CSV: %w", err)
	}

	fileName := fmt.Sprintf("dd-mandate-schedule-check/%s.csv", time.Now().UTC().Format("2006-01-02T15-04-05Z"))

	//The service builds the CSV in memory and uploads it:
	versionID, err := s.fileStorage.StreamFile(ctx, s.env.AsyncBucket, fileName, io.NopCloser(bytes.NewReader(csvBuffer.Bytes())))
	if err != nil {
		return fmt.Errorf("upload CSV to S3: %w", err)
	}

	logger.Info("DD mandate & schedule check: CSV uploaded", "bucket", s.env.AsyncBucket, "key", fileName, "versionId", versionID)
	return nil
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

		//based on the assumption that a court ref can only have one mandate
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

	// If either Allpay lookup failed, include the row so the error is visible in the CSV for investigation.
	if result.MandateError != "" || result.ScheduleError != "" {
		return true
	}

	// We are only interested in clients where Allpay has a mandate but no schedule.
	mandateExists := result.Mandate != nil && result.Mandate.TotalRecords > 0
	scheduleMissing := result.Schedule != nil && result.Schedule.TotalRecords == 0

	return mandateExists && scheduleMissing
}
