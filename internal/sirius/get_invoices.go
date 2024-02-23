package sirius

import (
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/internal/model"
	"net/http"
)

func (c *ApiClient) GetInvoices(ctx Context, ClientId int) (model.InvoiceList, error) {
	var invoiceList model.InvoiceList

	url := fmt.Sprintf("/api/v1/clients/%d/invoices", ClientId)

	req, err := c.newRequest(ctx, http.MethodGet, url, nil)

	if err != nil {
		return invoiceList, err
	}

	resp, err := c.http.Do(req)

	if err != nil {
		return invoiceList, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return invoiceList, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return invoiceList, newStatusError(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(&invoiceList); err != nil {
		return invoiceList, err
	}

	var invoices []model.Invoice

	for _, t := range invoiceList.Invoices {
		var invoice = model.Invoice{
			Id:                 t.Id,
			Ref:                t.Ref,
			Status:             t.Status,
			Amount:             t.Amount,
			RaisedDate:         t.RaisedDate,
			Received:           t.Received,
			OutstandingBalance: t.OutstandingBalance,
			Ledgers:            t.Ledgers,
			SupervisionLevels:  t.SupervisionLevels,
		}
		invoices = append(invoices, invoice)
	}

	invoiceList.Invoices = invoices
	//see returning just one odject not an array
	return invoiceList, err
}
