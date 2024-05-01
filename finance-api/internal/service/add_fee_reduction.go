package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/finance-api/internal/validation"
	"github.com/opg-sirius-finance-hub/shared"
	"strconv"
	"time"
)

func (s *Service) AddFeeReduction(body shared.AddFeeReduction) (shared.ValidationError, error) {
	ctx := context.Background()

	vError := validation.New()                                                        //move this up to the service so it will be s.vaildation
	vError.RegisterValidation("check-character-limit", vError.ValidateCharacterCount) //move to main
	vError.RegisterValidation("check-in-the-past", vError.ValidateDateInThePast)      //move to main

	checkForOverlappingFeeReductionParams := store.CheckForOverlappingFeeReductionParams{
		ClientID:  int32(body.ClientId),
		Startdate: calculateStartDate(body.StartYear),
		Enddate:   calculateEndDate(body.StartYear, body.LengthOfAward),
	}

	hasFeeReduction, err := s.Store.CheckForOverlappingFeeReduction(ctx, checkForOverlappingFeeReductionParams)
	if hasFeeReduction == 1 {
		var validationError shared.ValidationError
		var validationErrors = make(shared.ValidationErrors)
		if validationErrors["Overlap"] == nil {
			validationErrors["Overlap"] = make(map[string]string)
		}
		validationErrors["Overlap"]["StartOrEndDate"] = "A fee reduction already exists for the period specified"
		validationError.Message = "This has not worked"
		validationError.Errors = validationErrors
		return validationError, err
	}

	validationError := vError.ValidateStruct(body)

	if len(validationError.Errors) > 0 {
		return validationError, nil
	}

	queryArgs := store.AddFeeReductionParams{
		ClientID:     int32(body.ClientId),
		Type:         body.FeeType,
		Evidencetype: pgtype.Text{},
		Startdate:    calculateStartDate(body.StartYear),
		Enddate:      calculateEndDate(body.StartYear, body.LengthOfAward),
		Notes:        body.FeeReductionNotes,
		Deleted:      false,
		Datereceived: pgtype.Date{Time: body.DateReceive.Time, Valid: true},
	}

	_, err = s.Store.AddFeeReduction(ctx, queryArgs)
	if err != nil {
		return shared.ValidationError{}, err
	}

	return shared.ValidationError{}, nil
}

func calculateEndDate(startYear string, lengthOfAward string) pgtype.Date {
	startYearTransformed, _ := strconv.Atoi(startYear)
	lengthOfAwardTransformed, _ := strconv.Atoi(lengthOfAward)

	endDate := startYearTransformed + lengthOfAwardTransformed
	financeYearEnd := "-03-31"
	feeReductionEndDateString := strconv.Itoa(endDate) + financeYearEnd
	feeReductionEndDate, _ := time.Parse("2006-01-02", feeReductionEndDateString)
	return pgtype.Date{Time: feeReductionEndDate, Valid: true}
}

func calculateStartDate(startYear string) pgtype.Date {
	financeYearStart := "-04-01"
	feeReductionStartDateString := startYear + financeYearStart
	feeReductionStartDate, _ := time.Parse("2006-01-02", feeReductionStartDateString)
	return pgtype.Date{Time: feeReductionStartDate, Valid: true}
}
