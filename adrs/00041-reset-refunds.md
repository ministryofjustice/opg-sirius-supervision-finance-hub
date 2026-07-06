# 41. Reset Approved Refunds When Debt Changes

Date: 2026-06-03

## Status

Accepted

## Context

While the refund approval process is designed to be quick and simple, the business implementation of the process can take
a week or more between steps, due to the requirement for manager approval and Billing only downloading the approved refunds
file once a week. This had led to complications where the credit on the account changed between the approval and the refund
creating a ledger, which means a) more money is refunded than is available as credit, and b) the refund ledger is created
in an invalid state, where excess credit cannot be removed.

## Decision

To resolve this, we will reset the approved refunds to pending when the debt on the client changes. We could make this more
nuanced by only resetting refunds when the refunded amount exceeds available credit, but a refund that doesn't completely
cover excess credit would likely be rejected anyway, so this is a simpler solution.

A task will be created to review the refund.

## Consequences

There is still the possibility that the debt on a client could change between downloading the approved refunds file and
uploading the fulfilled refunds to create the ledger. However, this is done by Billing on the same day, and they would be
able to avoid downloading the report on days where large debt changes are likely to occur, such as annual billing. Additionally,
a refund is in an indeterminate state once the report has been downloaded, so the most we would be able to do is create
a task to notify the Billing team, which we can do at a later date if required.
