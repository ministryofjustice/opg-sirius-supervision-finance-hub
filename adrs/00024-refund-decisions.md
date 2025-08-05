# 24. Refund decisions

Date: 2025-06-12

## Status

Accepted

## Context

Like invoice adjustments, refunds are first proposed by a user as a pending refund, before a manager approves or rejects
it. However, refunds differs in that there are multiple stages of status changes after approval. This means that we can no
longer use the `status` field as an immutable value for the purpose of generating an audit trail for the billing history.

The statuses are as follows:
* PENDING - default state on creation
* REJECTED - the manager rejects the refund request (end of process)
* APPROVED - the manager approves the refund request
* PROCESSING - the refund report has been run and await confirmation from Bankline
* CANCELLED - an approved refund is no longer required (end of process)
* FULFILLED - refund is confirmed and ledger created (end of process)

## Decision

The following changes have been made:
* status column renamed to decision and only records PENDING, REJECTED, and APPROVED statuses
* processed_at timestamp = PROCESSING
* cancelled_at timestamp = CANCELLED
* fulfilled_at timestamp = FULFILLED


As the refund table does contain a link to the ledger it results in (is this true? Where is the link?), we could use the date on the ledger for the fulfilled
status, but as the refunds list needs to include the fulfilled date, it makes sense to duplicate it.

## Consequences

None.
