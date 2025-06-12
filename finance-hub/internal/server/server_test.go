package server

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"io"
	"net/http"
)

type mockTemplate struct {
	executed         bool
	executedTemplate bool
	lastVars         interface{}
	lastW            io.Writer
	error            error
}

func (m *mockTemplate) Execute(w io.Writer, vars any) error {
	m.executed = true
	m.lastVars = vars
	m.lastW = w
	return m.error
}

func (m *mockTemplate) ExecuteTemplate(w io.Writer, name string, vars any) error {
	m.executedTemplate = true
	m.lastVars = vars
	m.lastW = w
	return m.error
}

type mockRoute struct {
	client   ApiClient
	data     any
	executed bool
	lastW    io.Writer
	error
}

func (r *mockRoute) Client() ApiClient {
	return r.client
}

func (r *mockRoute) execute(w http.ResponseWriter, req *http.Request, data any) error {
	r.executed = true
	r.lastW = w
	r.data = data
	return r.error
}

type mockApiClient struct {
	error              error
	CurrentUserDetails shared.User
	PersonDetails      shared.Person
	FeeReductions      shared.FeeReductions
	Invoices           shared.Invoices
	Invoice            shared.Invoice
	AccountInformation shared.AccountInformation
	invoiceAdjustments shared.InvoiceAdjustments
	refunds            shared.Refunds
	BillingHistory     []shared.BillingHistory
	adjustmentTypes    []shared.AdjustmentType
	User               shared.User
}

func (m mockApiClient) SubmitPaymentMethod(context context.Context, i int, s string) error {
	return m.error
}

func (m mockApiClient) SubmitDirectDebit(s string, s2 string, s3 string, s4 string) error {
	return m.error
}

func (m mockApiClient) GetUser(context context.Context, i int) (shared.User, error) {
	return m.User, m.error
}

func (m mockApiClient) GetBillingHistory(context context.Context, i int) ([]shared.BillingHistory, error) {
	return m.BillingHistory, m.error
}

func (m mockApiClient) AddManualInvoice(context context.Context, i int, s string, s2 *string, s3 *string, s4 *string, s5 *string, s6 *string, s7 *string) error {
	return m.error
}

func (m mockApiClient) GetPermittedAdjustments(context.Context, int, int) ([]shared.AdjustmentType, error) {
	return m.adjustmentTypes, nil
}

func (m mockApiClient) UpdatePendingInvoiceAdjustment(context context.Context, i int, i2 int, i3 string) error {
	return m.error
}

func (m mockApiClient) AddInvoiceAdjustment(context.Context, int, int, int, string, string, string, bool) error {
	return m.error
}

func (m mockApiClient) CancelFeeReduction(context context.Context, i int, i2 int, s string) error {
	return m.error
}

func (m mockApiClient) UpdateInvoice(context.Context, int, int, string, string, string) error {
	return m.error
}

func (m mockApiClient) AddFeeReduction(context.Context, int, string, string, string, string, string) error {
	return m.error
}

func (m mockApiClient) GetInvoices(context.Context, int) (shared.Invoices, error) {
	return m.Invoices, m.error
}

func (m mockApiClient) GetPersonDetails(context.Context, int) (shared.Person, error) {
	return m.PersonDetails, m.error
}

func (m mockApiClient) GetCurrentUserDetails(context.Context) (shared.User, error) {
	return m.CurrentUserDetails, m.error
}

func (m mockApiClient) GetFeeReductions(context.Context, int) (shared.FeeReductions, error) {
	return m.FeeReductions, m.error
}

func (m mockApiClient) GetAccountInformation(context.Context, int) (shared.AccountInformation, error) {
	return m.AccountInformation, m.error
}

func (m mockApiClient) GetInvoiceAdjustments(context.Context, int) (shared.InvoiceAdjustments, error) {
	return m.invoiceAdjustments, m.error
}

func (m mockApiClient) GetRefunds(ctx context.Context, id int) (shared.Refunds, error) {
	return m.refunds, m.error
}

func (m mockApiClient) AddRefund(context.Context, int, string, string, string, string) error {
	return m.error
}

func (m mockApiClient) UpdateRefundStatus(context.Context, int, int, string) error {
	return m.error
}
