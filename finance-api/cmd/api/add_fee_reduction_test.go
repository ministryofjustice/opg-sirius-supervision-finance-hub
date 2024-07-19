package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/opg-sirius-finance-hub/finance-api/internal/service"
	"github.com/opg-sirius-finance-hub/finance-api/internal/validation"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServer_addFeeReductions(t *testing.T) {
	var b bytes.Buffer
	var dateReceivedTransformed *shared.Date

	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)
	dateReceivedTransformed = &shared.Date{Time: date}

	feeReductionInfo := &shared.AddFeeReduction{
		FeeType:       shared.FeeReductionTypeRemission,
		StartYear:     "2022",
		LengthOfAward: 1,
		DateReceived:  dateReceivedTransformed,
		Notes:         "Adding a remission reduction",
	}
	_ = json.NewEncoder(&b).Encode(feeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{feeReduction: feeReductionInfo}
	server := Server{Service: mock, Validator: validator}
	server.addFeeReduction(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := ""

	assert.Equal(t, expected, string(data))
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestServer_addFeeReductionsValidationErrors(t *testing.T) {
	var b bytes.Buffer
	feeReductionInfo := &shared.AddFeeReduction{
		FeeType:       shared.FeeReductionTypeUnknown,
		StartYear:     "",
		LengthOfAward: 0,
		DateReceived:  nil,
		Notes:         "",
	}
	_ = json.NewEncoder(&b).Encode(feeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{feeReduction: feeReductionInfo}
	server := Server{Service: mock, Validator: validator}
	server.addFeeReduction(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := `
{"Message":"","validation_errors":{"DateReceived":{"required":"This field DateReceived needs to be looked at required"},"FeeType":{"required":"This field FeeType needs to be looked at required"},"LengthOfAward":{"required":"This field LengthOfAward needs to be looked at required"},"Notes":{"required":"This field Notes needs to be looked at required"},"StartYear":{"required":"This field StartYear needs to be looked at required"}}}`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestServer_addFeeReductionsValidationErrorsForThousandCharacters(t *testing.T) {
	var b bytes.Buffer
	var dateReceivedTransformed *shared.Date

	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)
	dateReceivedTransformed = &shared.Date{Time: date}
	feeReductionInfo := &shared.AddFeeReduction{
		FeeType:       shared.FeeReductionTypeRemission,
		StartYear:     "2024",
		LengthOfAward: 1,
		DateReceived:  dateReceivedTransformed,
		Notes: "wC6fABXtm7LvSQ8oa3HUKsdtZldvEuvRwfyEKkAp8RsCxHWQjT8sWfj6cS1NzKVpG8AfAQ507IQ6zfKol" +
			"asWQ84zz6MzTLVbkXCbKWqx9jIsJn3klFGq4Q32O62FpiIsMUuJoGV1BsWFT9d9prh0sDIpyXTPdgXwCTL4iIAdydqpGlmHt" +
			"5dhyD4ZFYZICH2VFEWnTSrCGbBWPfbHArXxqZRCADf5ut3htEncnu0KSfSJhU2lSbT8erAueypq5u0Aot6fR0LKvtGuuK1VH" +
			"iEOaEayIcOZaZLxi9xRcXryW8weyIcw4FEWlBvxsN3ZtA1J94LQM4U41NdsZ18bzZrkQW3MFL8JOzgESIsjoxwqSDeTVuYgT" +
			"fkVdZcasrq0ao78jOq1ozvwJ3MKrbrOim10dmhmbkQlVCuEKKlt2HpgmpjC3CJRBRgNtYkdRAAcd8rgzjJxnMAIQwzwJ3Zw4" +
			"lik4P2ZINcMiQucpvAm4O4GhWwj6l0mcbjdNQT4n0MFIAV3HgbdZ6DfdR51urDrTxys5sjRMRbK4G8ida2ROMPy8ydnl96ut" +
			"nvIjjiLYfPzZVqcoUxJ34omPuXFpKsHXPJTplZrIQdGyeYJ3MGTyZFOG9Q9dGXwnyorjyzsyeH165uQgxPIsTmbrc3VjKjhF" +
			"LFvvNhUhjc9POyAOKnqP5YEEOWv7ubqXoU62gq4SijO4Ui8D1pnWRGlWGGLKDAkE9g9C3vzoBF542fdUDEu1URanf5dAQl9c" +
			"K1vfiPDdM6m9J2WAI7ReXHHW3cnTgkpLW2aHVhrU9ZkXgrMYgvBFC94W5jf19JsGnYlJrtEG37LuRdVwrc7jawzogffrwZVm" +
			"r5cobstMXqQBOWm18AwXVZJBk6aGmcTBTy0yzkqoqVfRFZ4mh9PScW7LYVdfNVFRa8agDiQOFqSuj8zrA89yufjO0Zube4wd" +
			"Sn3qgFi4p7hZJiFEIvvM1Xad9DA8H6KGFejzaBXZgkBuqY5duIjCRkADo",
	}
	_ = json.NewEncoder(&b).Encode(feeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{feeReduction: feeReductionInfo}
	server := Server{Service: mock, Validator: validator}
	server.addFeeReduction(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := `
{"Message":"","validation_errors":{"Notes":{"thousand-character-limit":"This field Notes needs to be looked at thousand-character-limit"}}}`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestServer_addFeeReductionsOverlapError(t *testing.T) {
	var b bytes.Buffer
	var dateReceivedTransformed *shared.Date

	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)
	dateReceivedTransformed = &shared.Date{Time: date}
	feeReductionInfo := &shared.AddFeeReduction{
		FeeType:       shared.FeeReductionTypeRemission,
		StartYear:     "2022",
		LengthOfAward: 1,
		DateReceived:  dateReceivedTransformed,
		Notes:         "Adding a remission reduction",
	}
	_ = json.NewEncoder(&b).Encode(feeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{feeReduction: feeReductionInfo, err: service.BadRequest{Reason: "overlap"}}
	server := Server{Service: mock, Validator: validator}
	server.addFeeReduction(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServer_addFeeReductions500Error(t *testing.T) {
	var b bytes.Buffer
	var dateReceivedTransformed *shared.Date

	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)
	dateReceivedTransformed = &shared.Date{Time: date}
	feeReductionInfo := &shared.AddFeeReduction{
		FeeType:       shared.FeeReductionTypeRemission,
		StartYear:     "2022",
		LengthOfAward: 1,
		DateReceived:  dateReceivedTransformed,
		Notes:         "Adding a remission reduction",
	}
	_ = json.NewEncoder(&b).Encode(feeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions", &b)
	req.SetPathValue("clientId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{feeReduction: feeReductionInfo, err: errors.New("Something is wrong")}
	server := Server{Service: mock, Validator: validator}
	server.addFeeReduction(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
