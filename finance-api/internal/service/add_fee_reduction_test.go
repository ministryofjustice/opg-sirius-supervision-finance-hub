package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/finance-api/internal/validation"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func setupServiceAndParams() (*Service, shared.AddFeeReduction) {
	Store := store.New(testDB.DbInstance)
	//DB := testDB.DbInstance
	vError := validation.New()
	vError.RegisterValidation("thousand-character-limit", vError.ValidateThousandCharacterCount)
	vError.RegisterValidation("date-in-the-past", vError.ValidateDateInThePast)

	today := time.Now()
	dateInRangeOfSixMonths := today.AddDate(0, -5, -29).Format("2006-01-02")
	dateInRangeOfSixMonthsToTime, _ := time.Parse("2006-01-02", dateInRangeOfSixMonths)

	params := shared.AddFeeReduction{
		ClientId:          5,
		FeeType:           "remission",
		StartYear:         "2021-04-01",
		LengthOfAward:     "3",
		DateReceive:       shared.Date{Time: dateInRangeOfSixMonthsToTime},
		FeeReductionNotes: "Testing",
	}

	s := &Service{
		Store:  Store,
		DB:     testDB.DbConn,
		VError: vError,
	}

	return s, params
}

func TestService_AddFeeReduction(t *testing.T) {
	s, params := setupServiceAndParams()

	got, _ := s.AddFeeReduction(params)
	if len(got.Errors) == 0 {
		return
	}
}

func TestService_AddFeeReductionErrorForRequiredFields(t *testing.T) {
	s, params := setupServiceAndParams()
	params.ClientId = 0
	params.FeeType = ""
	params.StartYear = ""
	params.LengthOfAward = ""
	params.DateReceive = shared.Date{}
	params.FeeReductionNotes = ""

	got, _ := s.AddFeeReduction(params)

	expectedErrorCount := 5
	if len(got.Errors) != expectedErrorCount {
		t.Errorf("AddFeeReduction() returned unexpected number of validation errors: got %d, want %d", len(got.Errors), expectedErrorCount)
	}

	expectedErrors := map[string]string{
		"FeeReductionNotes": "This field FeeReductionNotes needs to be looked at required",
		"FeeType":           "This field FeeType needs to be looked at required",
		"LengthOfAward":     "This field LengthOfAward needs to be looked at required",
		"StartYear":         "This field StartYear needs to be looked at required",
	}

	for field, expectedMessage := range expectedErrors {
		actualMessage, ok := got.Errors[field]
		if !ok {
			t.Errorf("AddFeeReduction() missing expected error for field %s", field)
		}
		if actualMessage["required"] != expectedMessage {
			t.Errorf("AddFeeReduction() returned unexpected error message for field %s: got %s, want %s", field, actualMessage[field], expectedMessage)
		}
	}
}

func TestService_AddFeeReductionErrorForOver1000CharactersFields(t *testing.T) {
	s, params := setupServiceAndParams()
	params.FeeReductionNotes = "wC6fABXtm7LvSQ8oa3HUKsdtZldvEuvRwfyEKkAp8RsCxHWQjT8sWfj6cS1NzKVpG8AfAQ507IQ6zfKol" +
		"asWQ84zz6MzTLVbkXCbKWqx9jIsJn3klFGq4Q32O62FpiIsMUuJoGV1BsWFT9d9prh0sDIpyXTPdgXwCTL4iIAdydqpGlmHt" +
		"5dhyD4ZFYZICH2VFEWnTSrCGbBWPfbHArXxqZRCADf5ut3htEncnu0KSfSJhU2lSbT8erAueypq5u0Aot6fR0LKvtGuuK1VH" +
		"iEOaEayIcOZaZLxi9xRcXryW8weyIcw4FEWlBvxsN3ZtA1J94LQM4U41NdsZ18bzZrkQW3MFL8JOzgESIsjoxwqSDeTVuYgT" +
		"fkVdZcasrq0ao78jOq1ozvwJ3MKrbrOim10dmhmbkQlVCuEKKlt2HpgmpjC3CJRBRgNtYkdRAAcd8rgzjJxnMAIQwzwJ3Zw4" +
		"lik4P2ZINcMiQucpvAm4O4GhWwj6l0mcbjdNQT4n0MFIAV3HgbdZ6DfdR51urDrTxys5sjRMRbK4G8ida2ROMPy8ydnl96ut" +
		"nvIjjiLYfPzZVqcoUxJ34omPuXFpKsHXPJTplZrIQdGyeYJ3MGTyZFOG9Q9dGXwnyorjyzsyeH165uQgxPIsTmbrc3VjKjhF" +
		"LFvvNhUhjc9POyAOKnqP5YEEOWv7ubqXoU62gq4SijO4Ui8D1pnWRGlWGGLKDAkE9g9C3vzoBF542fdUDEu1URanf5dAQl9c" +
		"K1vfiPDdM6m9J2WAI7ReXHHW3cnTgkpLW2aHVhrU9ZkXgrMYgvBFC94W5jf19JsGnYlJrtEG37LuRdVwrc7jawzogffrwZVm" +
		"r5cobstMXqQBOWm18AwXVZJBk6aGmcTBTy0yzkqoqVfRFZ4mh9PScW7LYVdfNVFRa8agDiQOFqSuj8zrA89yufjO0Zube4wd" +
		"Sn3qgFi4p7hZJiFEIvvM1Xad9DA8H6KGFejzaBXZgkBuqY5duIjCRkADo"

	got, _ := s.AddFeeReduction(params)

	expectedErrorCount := 1
	if len(got.Errors) != expectedErrorCount {
		t.Errorf("AddFeeReduction() returned unexpected number of validation errors: got %d, want %d", len(got.Errors), expectedErrorCount)
	}

	expectedErrors := map[string]string{
		"FeeReductionNotes": "This field FeeReductionNotes needs to be looked at thousand-character-limit",
	}

	for field, expectedMessage := range expectedErrors {
		actualMessage, ok := got.Errors[field]
		if !ok {
			t.Errorf("AddFeeReduction() missing expected error for field %s", field)
		}
		if actualMessage["thousand-character-limit"] != expectedMessage {
			t.Errorf("AddFeeReduction() returned unexpected error message for field %s: got %s, want %s", field, actualMessage[field], expectedMessage)
		}
	}
}

func TestService_AddFeeReductionErrorForDateNotInThePastsFields(t *testing.T) {
	s, params := setupServiceAndParams()
	params.DateReceive.Time = time.Now().AddDate(0, 0, 1)

	got, _ := s.AddFeeReduction(params)

	expectedErrorCount := 1
	if len(got.Errors) != expectedErrorCount {
		t.Errorf("AddFeeReduction() returned unexpected number of validation errors: got %d, want %d", len(got.Errors), expectedErrorCount)
	}

	expectedErrors := map[string]string{
		"DateReceive": "This field DateReceive needs to be looked at date-in-the-past",
	}

	for field, expectedMessage := range expectedErrors {
		actualMessage, ok := got.Errors[field]
		if !ok {
			t.Errorf("AddFeeReduction() missing expected error for field %s", field)
		}
		if actualMessage["date-in-the-past"] != expectedMessage {
			t.Errorf("AddFeeReduction() returned unexpected error message for field %s: got %s, want %s", field, actualMessage[field], expectedMessage)
		}
	}
}

func Test_calculateEndDate(t *testing.T) {
	type args struct {
		startYear     string
		lengthOfAward string
	}
	tests := []struct {
		name string
		args args
		want pgtype.Date
	}{
		{
			name: "returns the correct end date for a three year length award",
			args: args{
				startYear:     "2024",
				lengthOfAward: "3",
			},
			want: pgtype.Date{Time: time.Date(2027, time.March, 31, 0, 0, 0, 0, time.UTC), Valid: true},
		},
		{
			name: "returns the correct end date for a two year length award",
			args: args{
				startYear:     "2024",
				lengthOfAward: "2",
			},
			want: pgtype.Date{Time: time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC), Valid: true},
		},
		{
			name: "returns the correct end date for a one year length award",
			args: args{
				startYear:     "2024",
				lengthOfAward: "1",
			},
			want: pgtype.Date{Time: time.Date(2025, time.March, 31, 0, 0, 0, 0, time.UTC), Valid: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, calculateEndDate(tt.args.startYear, tt.args.lengthOfAward), "calculateEndDate(%v, %v)", tt.args.startYear, tt.args.lengthOfAward)
		})
	}
}

func Test_calculateStartDate(t *testing.T) {
	type args struct {
		startYear string
	}
	tests := []struct {
		name string
		args args
		want pgtype.Date
	}{
		{
			name: "returns the correct start date",
			args: args{
				startYear: "2024",
			},
			want: pgtype.Date{Time: time.Date(2024, time.April, 01, 0, 0, 0, 0, time.UTC), Valid: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, calculateStartDate(tt.args.startYear), "calculateStartDate(%v)", tt.args.startYear)
		})
	}
}
