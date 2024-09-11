package api

//
//func TestServer_getBillingHistory(t *testing.T) {
//	req := httptest.NewRequest(http.MethodGet, "/clients/1/billing-history", nil)
//	req.SetPathValue("clientId", "1")
//	w := httptest.NewRecorder()
//
//	billingHistoryInfo := []shared.BillingHistory{
//		{
//			User: 65,
//			Date: shared.Date{Time: time.Date(2099, time.November, 4, 15, 4, 5, 0, time.UTC)},
//			Event: &shared.InvoiceAdjustmentPending{
//				AdjustmentType: shared.AdjustmentTypeWriteOff,
//				ClientId:       1,
//				PaymentBreakdown: shared.PaymentBreakdown{
//					InvoiceReference: shared.InvoiceEvent{
//						ID:        1,
//						Reference: "S203531/19",
//					},
//					Amount: 12300,
//				},
//				BaseBillingEvent: shared.BaseBillingEvent{Type: shared.EventTypeInvoiceAdjustmentPending},
//			},
//			OutstandingBalance: 0,
//		},
//	}
//
//	mock := &mockService{billingHistory: billingHistoryInfo}
//	server := Server{Service: mock}
//	server.getBillingHistory(w, req)
//
//	res := w.Result()
//	defer res.Body.Close()
//	data, _ := io.ReadAll(res.Body)
//
//	expected := `[{"user":65,"date":"04\/11\/2099","event":{"adjustment_type":"CREDIT WRITE OFF","client_id":1,"notes":"","payment_breakdown":{"invoice_reference":{"id":1,"reference":"S203531/19"},"amount":12300},"type":"INVOICE_ADJUSTMENT_PENDING"},"outstanding_balance":0}]`
//
//	assert.Equal(t, expected, string(data))
//	assert.Equal(t, 1, mock.expectedIds[0])
//	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
//}
//
//func TestServer_getBillingHistory_returns_an_empty_array(t *testing.T) {
//	req := httptest.NewRequest(http.MethodGet, "/clients/2/billing-history", nil)
//	req.SetPathValue("clientId", "2")
//	w := httptest.NewRecorder()
//
//	mock := &mockService{billingHistory: []shared.BillingHistory{}}
//	server := Server{Service: mock}
//	server.getBillingHistory(w, req)
//
//	res := w.Result()
//	defer res.Body.Close()
//	data, _ := io.ReadAll(res.Body)
//
//	expected := `[]`
//
//	assert.Equal(t, expected, string(data))
//	assert.Equal(t, 2, mock.expectedIds[0])
//	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
//}
//
//func TestServer_getBillingHistory_error(t *testing.T) {
//	req := httptest.NewRequest(http.MethodGet, "/clients/1/billing-history", nil)
//	req.SetPathValue("clientId", "1")
//	w := httptest.NewRecorder()
//
//	mock := &mockService{err: pgx.ErrTooManyRows}
//	server := Server{Service: mock}
//	server.getInvoices(w, req)
//
//	res := w.Result()
//
//	assert.Equal(t, 500, res.StatusCode)
//}
