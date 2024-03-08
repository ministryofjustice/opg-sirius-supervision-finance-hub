package server

import (
	"github.com/opg-sirius-finance-hub/shared"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"net/http"
)

type FeeReductions []FeeReduction

type FeeReduction struct {
	Type         string
	StartDate    string
	EndDate      string
	DateReceived string
	Notes        string
	Status       string
}

type FeeReductionsTab struct {
	FeeReductions FeeReductions
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

	data := FeeReductionsTab{h.transform(feeReductions), v}
	return h.execute(w, r, data)
}

func (h *FeeReductionsHandler) transform(in shared.FeeReductions) FeeReductions {
	var out FeeReductions
	for _, f := range in {
		out = append(out, FeeReduction{
			Type:         cases.Title(language.English).String(f.Type),
			StartDate:    f.StartDate.String(),
			EndDate:      f.EndDate.String(),
			DateReceived: f.DateReceived.String(),
			Notes:        f.Notes,
		})
	}
	return out
}
