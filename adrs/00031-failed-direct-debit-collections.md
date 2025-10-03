# 31. Failed Direct Debit collections

Date: 2025-09-30

## Status

Accepted

## Context

Failed Direct Debit collections can be obtained by calling the Allpay API with a date range based on the collection date
of the payment. Due to the length of time it takes BACS to process payments, a bank account is not credited with the reversed
payment until up to six days from the collection date. Therefore, we need to be able to poll for this information in a way
that neither misses failed payments, nor double counts ones that have already been processed.

We have two options:
1. Fetch failed payments for a collection date seven working days in the past
2. Fetch for a seven working day window on a rolling nightly basis

## Decision

We will use the rolling window approach. This provides more flexibility and does not rely on any guarantees that the data
is immutable after a certain date.

As the original ledgers are created on the collection date and the reversal is credited to the bank account on the processed
date, we will use those dates to match the ledger and create the reversal, i.e. the find the ledger to reverse using the
collection date as the bank and received dates, and use the processed date for those dates in the reversal ledger.

## Consequences

As these are working days, we cannot use the EventBridge scheduler to populate those date ranges with cron, so will need
to calculate those when building the API request.

With a seven working day window, we should be able to capture all failed payments. If any are missed, a date override has
been built into the event, so we can manually trigger it if needed.

As this is an automated task, we can't send an email of failed rows as we do with uploaded payment files. However, if we 
can assume that all Direct Debit collections that could fail were created via automated processes, i.e. through our prior
interactions with the API, any error should be a system failure rather than user error, so logging and alerting are the 
best options.
