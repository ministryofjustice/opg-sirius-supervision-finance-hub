package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) GetAnnualBillingInfo(ctx context.Context) (shared.AnnualBillingInformation, error) {
	annualBillingYear, err := s.store.GetAnnualBillingYear(ctx)
	if err != nil {
		return shared.AnnualBillingInformation{}, err
	}
	//leave for testing remove and add error handling for live
	if annualBillingYear == "" {
		annualBillingYear = "2025"
	}

	billingYearStartDate := annualBillingYear + "-04-01"
	thisYear, _ := strconv.Atoi(annualBillingYear)
	nextYear := thisYear + 1
	billingYearEndDate := strconv.Itoa(nextYear) + "-03-31"
	start, _ := time.Parse("2006-01-02", billingYearStartDate)
	end, _ := time.Parse("2006-01-02", billingYearEndDate)
	startDate := pgtype.Date{Time: start, Valid: true}
	endDate := pgtype.Date{Time: end, Valid: true}

	info, err := s.store.GetAnnualBillingLettersInformation(ctx, store.GetAnnualBillingLettersInformationParams{Column1: startDate, Column2: endDate})
	if err != nil {
		return shared.AnnualBillingInformation{}, err
	}

	fmt.Println("info:")
	response := shared.AnnualBillingInformation{
		AnnualBillingYear: annualBillingYear,
	}

	for i := range info {
		switch info[i].Status.String {
		case "UNPROCESSED":
			response.ExpectedCount = info[i].Count.Int64
		case "SKIPPED":
			response.SkippedCount = info[i].Count.Int64
		default:
			response.IssuedCount = response.IssuedCount + info[i].Count.Int64
		}
	}
	return response, nil
}
