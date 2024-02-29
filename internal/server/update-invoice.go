package server

import (
	"fmt"
	"github.com/opg-sirius-finance-hub/internal/model"
	"net/http"
)

type UpdateInvoices struct {
	InvoiceTypes []model.InvoiceType
	InvoiceType  string
	Notes        string
	Id           string
	AppVars
}

type UpdateInvoiceHandler struct {
	router
}

func (h *UpdateInvoiceHandler) render(v AppVars, w http.ResponseWriter, r *http.Request) error {
	invoiceTypes := []model.InvoiceType{
		{Handle: "writeOff", Description: "Write off"},
		{Handle: "addCredit", Description: "Add credit"},
		{Handle: "addDebit", Description: "Add debit"},
		{Handle: "unapply", Description: "Unapply"},
		{Handle: "reapply", Description: "Reapply"},
	}
	data := UpdateInvoices{invoiceTypes, "", "", r.PathValue("id"), v}

	if r.Method == http.MethodGet {
		return h.execute(w, r, data)
	} else {
		fmt.Println("other one")
		ctx := getContext(r)
		//var (
		//	invoiceId, _ = strconv.Atoi(r.PathValue("id"))
		//	invoiceType  = r.PostFormValue("invoiceType")
		//	//typeName     = getInvoiceName(invoiceType, invoiceTypes)
		//	notes = r.PostFormValue("notes")
		//)
		//
		//err := h.Client().UpdateInvoice(ctx, ctx.ClientId, invoiceId, typeName, notes)
		//var verr sirius.ValidationError
		//if errors.As(err, &verr) {
		//	data.InvoiceTypes = invoiceTypes
		//	data.InvoiceType = invoiceType
		//	data.Notes = notes
		//	data.Errors = util.RenameErrors(verr.Errors)
		//	w.WriteHeader(http.StatusBadRequest)
		//	return h.execute(w, r, data)
		//}
		//if err != nil {
		//	return err
		//}
		//
		//var invoiceName string
		//for _, t := range invoiceTypes {
		//	if t.Handle == invoiceType {
		//		invoiceName = t.Description
		//	}
		//}

		return RedirectError(fmt.Sprintf("%s/clients/%d/invoices", v.EnvironmentVars.Prefix, ctx.ClientId))
	}
}

//
//func getInvoiceName(handle string, types []model.InvoiceType) string {
//	for _, t := range types {
//		if handle == t.Handle {
//			return t.Description
//		}
//	}
//	return ""
//}
