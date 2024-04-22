package server

import (
	"github.com/opg-sirius-finance-hub/shared"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"net/http"
	"strconv"
)

type FeeReductions []FeeReduction

type FeeReduction struct {
	Type         string
	StartDate    string
	EndDate      string
	DateReceived string
	Status       string
	Notes        string
}

type FeeReductionsTab struct {
	FeeReductions   FeeReductions
	FinanceClientId string
	AppVars
}

type FeeReductionsHandler struct {
	router
}

func (h *FeeReductionsHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	ctx := getContext(r)

	feeReductions, err := h.Client().GetFeeReductions(ctx, ctx.ClientId)
	if err != nil {
		return err
	}

	feeReductionData, financeClientId := h.transform(feeReductions)

	data := &FeeReductionsTab{feeReductionData, financeClientId, v}
	data.selectTab("fee-reductions")
	return h.execute(w, r, data)
}

func (h *FeeReductionsHandler) transform(in shared.FeeReductions) (FeeReductions, string) {
	var out FeeReductions
	var financeClientId string
	for _, f := range in {
		out = append(out, FeeReduction{
			Type:         cases.Title(language.English).String(f.Type),
			StartDate:    f.StartDate.String(),
			EndDate:      f.EndDate.String(),
			DateReceived: f.DateReceived.String(),
			Status:       f.Status,
			Notes:        f.Notes,
		})
		financeClientId = strconv.Itoa(f.FinanceClientId)
	}
	return out, financeClientId
}
