package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
)

func (s *Service) GetAnnualBillingInfo(ctx context.Context) (map[string]interface{}, error) {
	annualBillingYear, err := s.store.GetAnnualBillingYear(ctx)
	if err != nil {
		return nil, err
	}
	//leave for testing remove and add error handling for live
	if annualBillingYear == "" {
		annualBillingYear = "2025"
	}

	billingYearStartDate := annualBillingYear + "-04-01"

	thisYear, err := strconv.Atoi(annualBillingYear)
	nextYear := thisYear + 1
	billingYearEndDate := strconv.Itoa(nextYear) + "-03-31"
	start, _ := time.Parse("2006-01-02", billingYearStartDate)
	end, _ := time.Parse("2006-01-02", billingYearEndDate)
	startDate := pgtype.Date{Time: start, Valid: true}
	endDate := pgtype.Date{Time: end, Valid: true}

	info, err := s.store.GetAnnualBillingLettersInformation(ctx, store.GetAnnualBillingLettersInformationParams{Column1: startDate, Column2: endDate})
	if err != nil {
		return nil, err
	}

	fmt.Println(info)
	response := make(map[string]interface{})

	response["expectedCount"] = 3
	response["issuedCount"] = 4
	response["skippedCount"] = 1

	return response, nil
}
