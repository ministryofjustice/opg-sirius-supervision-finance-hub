# 20. Payment reversals

Date: 2025-03-28

## Status

Accepted

## Context

Due to end user error it is possible for payments uploaded into Sirius to be allocated incorrectly. To rectify 
this the payment must be reversed against that client and the monies unallocated from them. This must be achieved by a 
payment reversal for accounting purposes so using equivalent credits/debits is not appropriate.

A "misapplied payment" is a payment that was allocated to the wrong client. To rectify this, the payment is first reversed
against the client it was allocated to, and then a new payment is created for the correct client, using the same transaction
dates.

## Decision

Payment reversals are implemented as duplicates of the original payment, but with a negative amount. This is in contrast
with the alternative, which would be to have a reverse of the ledger. The payment option was chosen as the simpler option,
as a ledger's content is dependent on the financial picture at creation, which will be different when the payment is reversed.
For instance, a ledger may result in an overpayment, but that credit may have been reapplied at the point of reversal, or
another payment was received in the meantime. This could lead to a situation where the most recent invoice is paid off but
payments are removed from older invoices, and as payments are taken for a client and not for a particular invoice, this 
would be a departure from our financial processes.

## Consequences

This required a change in many of the reports, schedules, and journals, as the assumption up until now was that all 
payments would be positive values.
