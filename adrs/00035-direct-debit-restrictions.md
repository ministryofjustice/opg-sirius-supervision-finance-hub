# 35. Direct Debit API restrictions

Date: 2026-04-13

## Status

Accepted

## Context

When the integration between Allpay, Payments, and Sirius was first implemented, the services were kept agnostic of their
producers and consumers. For example, if Payments receives an event to close a Direct Debit, it would attempt to close it
without being concerned as to why the event was triggered. Under event-driven architecture, this is correct from the perspective
of the producer, but the consumer need more control of which events are processed, e.g. Sirius as the producer should broadcast 
an event when the client is made inactive, but Payments as the consumer should check the client has a mandate before attempting 
to close it.

There was also an assumption that while events could be sent and received as a result of plausible actions, these actions 
would never be possible due to other business rules. For instance, while 'invoice-created' events are sent for all invoices
created in Sirius, AD fees are created before Direct Debit mandates are sent out, and final fees would be sent after the
mandate has been cancelled, so neither _should_ result in a payment schedule being created. However, business rules are
messier than expected and these unenforced restrictions can be bypassed, such as when an AD fee is created for a replacement
order and the client still has an active mandate.

## Decision

Event consumers within the Payments service will implement additional checks to ensure the actions they take in response
to events are appropriate. When receiving an event to close a Direct Debit, Payments will check if the client 
has their payment method set to 'Direct Debit' before attempting to close it. If the client does not have an active mandate, 
the event will be ignored. Similarly, when receiving an 'invoice-created' event, Payments will only attempt to create a 
payment schedule if the invoice type is B2/B3, i.e. an annual invoice paid by Direct Debit.

## Consequences

This brings us better in line with event-driven architecture best practice. However, it does mean both developers and producer
services need to be aware of the conditions the consumer (Payments) will check before taking action. For instance, if we
decide replacement AD fees should be paid by Direct Debit where an existing mandate exists, the developer will need to
understand where the accepted invoice types are hardcoded.
