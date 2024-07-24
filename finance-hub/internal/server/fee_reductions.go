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
	Type                     string
	StartDate                string
	EndDate                  string
	DateReceived             string
	Status                   string
	Notes                    string
	FeeReductionCancelAction bool
	Id                       string
}

type FeeReductionsTab struct {
	FeeReductions FeeReductions
	ClientId      string
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

	data := &FeeReductionsTab{h.transform(feeReductions), strconv.Itoa(ctx.ClientId), v}
	data.selectTab("fee-reductions")
	return h.execute(w, r, data)
}

func (h *FeeReductionsHandler) transform(in shared.FeeReductions) FeeReductions {
	var out FeeReductions
	caser := cases.Title(language.English)
	for _, f := range in {
		out = append(out, FeeReduction{
			Type:                     f.Type.String(),
			StartDate:                f.StartDate.String(),
			EndDate:                  f.EndDate.String(),
			DateReceived:             f.DateReceived.String(),
			Status:                   f.Status,
			Notes:                    f.Notes,
			FeeReductionCancelAction: showFeeReductionCancelBtn(caser.String(f.Status)),
			Id:                       strconv.Itoa(f.Id),
		})
	}
	return out
}

func showFeeReductionCancelBtn(status string) bool {
	if status == shared.StatusPending || status == shared.StatusActive {
		return true
	}
	return false
}
