package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"time"
)

func (s *Service) ProcessAdhocEvent(ctx context.Context) error {

	ctx, cancelTx := s.WithCancel(ctx)
	defer cancelTx()

	tx, err := s.BeginStoreTx(ctx)
	if err != nil {
		return err
	}

	invoices, err := tx.GetNegativeInvoices(ctx)
	if err != nil {
		return err
	}

	var clientIDs []int32

	for _, invoice := range invoices {
		if invoice.Type == "UNKNOWN DEBIT" {
			invoice.Type = "CREDIT REMISSION"
		}

		if invoice.Type == "UNKNOWN CREDIT" {
			if invoice.Ledgerid == 652949 || invoice.Ledgerid == 652948 {
				invoice.Type = "CREDIT REMISSION"
			} else {
				invoice.Type = "CREDIT EXEMPTION"
			}
		}

		now := time.Now().UTC()
		todaysDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		var (
			receivedDate pgtype.Timestamp
			courtRef     pgtype.Text
			billingDate  pgtype.Date
			createdBy    pgtype.Int4
			pisNumber    pgtype.Int4
		)

		_ = receivedDate.Scan(todaysDate)
		_ = courtRef.Scan(invoice.CourtRef.String)
		_ = billingDate.Scan(todaysDate)
		_ = createdBy.Scan(ctx.(auth.Context).User.ID)
		amount := int32(invoice.Ledgerallocationamountneeded) // #nosec G115 -- This is a false positive

		params := store.CreateLedgerForCourtRefParams{
			CourtRef:     courtRef,
			Amount:       amount,
			Type:         invoice.Type,
			Status:       "CONFIRMED",
			CreatedBy:    createdBy,
			BankDate:     billingDate,
			ReceivedDate: receivedDate,
			PisNumber:    pisNumber,
		}

		ledgerID, err := tx.CreateLedgerForCourtRef(ctx, params)
		if err != nil {
			return err
		}

		var invoiceID pgtype.Int4
		_ = store.ToInt4(&invoiceID, invoice.Invoiceid)

		err = tx.CreateLedgerAllocation(ctx, store.CreateLedgerAllocationParams{
			InvoiceID: invoiceID,
			Amount:    amount,
			Status:    "UNAPPLIED",
			LedgerID:  ledgerID,
		})
		if err != nil {
			return err
		}
		clientIDs = append(clientIDs, invoice.PersonID.Int32)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	for _, clientID := range clientIDs {
		err = s.ReapplyCredit(ctx, clientID)
		if err != nil {
			return err
		}
	}

	return nil
}
