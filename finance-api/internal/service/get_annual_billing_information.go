package service

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (s *Service) GetAnnualBillingInformation(ctx context.Context) (shared.AnnualBillingInformation, error) {
	annualBillingYear, err := s.store.GetAnnualBillingYear(ctx)
	if err != nil {
		return shared.AnnualBillingInformation{}, err
	}

	annualBillingYearInt := int(annualBillingYear.Int32)
	billingYearStartDate := strconv.Itoa(annualBillingYearInt) + "-04-01"
	billingYearEndDate := strconv.Itoa(annualBillingYearInt+1) + "-03-31"
	var startDate, endDate pgtype.Date
	_ = startDate.Scan(billingYearStartDate)
	_ = endDate.Scan(billingYearEndDate)

	info, err := s.store.GetAnnualBillingLettersInformation(ctx, store.GetAnnualBillingLettersInformationParams{Startdate: startDate, Enddate: endDate})
	if err != nil {
		return shared.AnnualBillingInformation{}, err
	}

	response := shared.AnnualBillingInformation{
		AnnualBillingYear: annualBillingYearInt,
	}

	for _, el := range info {
		if el.PaymentMethod == "DIRECT DEBIT" {
			switch el.Status.String {
			case "UNPROCESSED":
				response.DirectDebitExpectedCount = int(el.Count.Int64)
			default:
				response.DirectDebitIssuedCount = response.DirectDebitIssuedCount + int(el.Count.Int64)
			}
		}
		if el.PaymentMethod == "DEMANDED" {
			switch el.Status.String {
			case "UNPROCESSED":
				response.DemandedExpectedCount = int(el.Count.Int64)
			case "SKIPPED":
				response.DemandedSkippedCount = int(el.Count.Int64)
			default:
				response.DemandedIssuedCount = response.DemandedIssuedCount + int(el.Count.Int64)
			}
		}
	}
	return response, nil
}
