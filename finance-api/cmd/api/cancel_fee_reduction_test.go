package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/opg-sirius-finance-hub/finance-api/internal/validation"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_cancelFeeReduction(t *testing.T) {
	var b bytes.Buffer

	cancelFeeReductionInfo := &shared.CancelFeeReduction{
		Notes: "Cancelling a remission reduction",
	}
	_ = json.NewEncoder(&b).Encode(cancelFeeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions/1/cancel", &b)
	req.SetPathValue("id", "1")
	req.SetPathValue("feeReductionId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{cancelFeeReduction: cancelFeeReductionInfo}
	server := Server{Service: mock, Validator: validator}
	server.cancelFeeReduction(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := ""

	assert.Equal(t, expected, string(data))
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestServer_cancelFeeReductionsValidationErrors(t *testing.T) {
	var b bytes.Buffer
	cancelFeeReductionInfo := &shared.CancelFeeReduction{
		Notes: "",
	}
	_ = json.NewEncoder(&b).Encode(cancelFeeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions/1/cancel", &b)
	req.SetPathValue("id", "1")
	req.SetPathValue("feeReductionId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{cancelFeeReduction: cancelFeeReductionInfo}
	server := Server{Service: mock, Validator: validator}
	server.cancelFeeReduction(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := `
{"Message":"","validation_errors":{"CancelFeeReductionNotes":{"required":"This field CancelFeeReductionNotes needs to be looked at required"}}}`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestServer_cancelFeeReductionsValidationErrorsForThousandCharacters(t *testing.T) {
	var b bytes.Buffer

	cancelFeeReductionInfo := &shared.CancelFeeReduction{
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
	_ = json.NewEncoder(&b).Encode(cancelFeeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions/1/cancel", &b)
	req.SetPathValue("id", "1")
	req.SetPathValue("feeReductionId", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{cancelFeeReduction: cancelFeeReductionInfo}
	server := Server{Service: mock, Validator: validator}
	server.cancelFeeReduction(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	expected := `
{"Message":"","validation_errors":{"CancelFeeReductionNotes":{"thousand-character-limit":"This field CancelFeeReductionNotes needs to be looked at thousand-character-limit"}}}`

	assert.Equal(t, expected, string(data))
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestServer_cancelFeeReductions500Error(t *testing.T) {
	var b bytes.Buffer
	cancelFeeReductionInfo := &shared.CancelFeeReduction{
		Notes: "Adding a remission reduction",
	}
	_ = json.NewEncoder(&b).Encode(cancelFeeReductionInfo)
	req := httptest.NewRequest(http.MethodPost, "/clients/1/fee-reductions/1/cancel", &b)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	validator, _ := validation.New()

	mock := &mockService{cancelFeeReduction: cancelFeeReductionInfo, err: errors.New("Something is wrong")}
	server := Server{Service: mock, Validator: validator}
	server.cancelFeeReduction(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
