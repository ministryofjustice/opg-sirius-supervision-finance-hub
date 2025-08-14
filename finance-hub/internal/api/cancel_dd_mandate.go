package api

import (
	"context"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

func (c *Client) CancelDirectDebitMandate(ctx context.Context, clientId int) error {
	logger := telemetry.LoggerFromContext(ctx)

	client, err := c.GetPersonDetails(ctx, clientId)
	if err != nil {
		return err
	}

	err = c.allpayClient.CancelMandate(ctx, &allpay.CancelMandateRequest{
		ClientReference: client.CourtRef,
		Surname:         client.Surname,
	})
	if err != nil {
		return err
	}

	err = c.UpdatePaymentMethod(ctx, clientId, shared.PaymentMethodDemanded.Key())
	if err != nil {
		logger.Error("failed to update payment method in Sirius after successful mandate cancellation in AllPay", "error", err)
		return err
	}
	return nil
}
