# 43. Existing Direct Debit Mandates

Date: 2026-07-30

## Status

Accepted

## Context

We have encountered a number of instances of mandates failing to be created because they already exist in Allpay. This
could be for a number of reasons, including:
* Migration issues when moving from SSCL to Allpay
* A mandate being cancelled and then recreated before the cancellation date had passed, due to a future pending collection
* A mandate being cancelled and then recreated the same day, in order to update bank details

While the first scenario is unavoidable if it occurs, and the third can be solved by updating business processes for Billing
team users, the second scenario is a legitimate use case that we need to support.

## Decision

Allpay returns a specific validation error when a mandate already exists, which we can use to identify this scenario and
handle it by proceeding as if mandate creation was successful. We then need to attempt to create the schedule, if there
was debt to collect as part of the original mandate request.

## Consequences

This does potentially re-introduce the same issue that we considered in ADR 42, where the Create Mandate request "succeeds"
(returns the correct validation error) but the schedule creation fails. However, this is a much more limited failure as it
only occurs where a mandate already exists for a client in Allpay. It would also be easier to identify, as both calls are
made within the same client request, so not only can the client retry the request via the UI, the payment method will not
have been updated to Direct Debit, which was an original cause of confusion for users in diagnosing the issue.
