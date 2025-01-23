# 17. Ledger status

Date: 2025-01-14

## Status

Accepted

## Context

Historically, ledgers created in Sirius (e.g. credits, fee reductions, etc.) needed to be confirmed by SOP before they were 
considered to apply to the debt position. Therefore, ledgers created in this way would begin in an `APPROVED` state before
transitioning to `CONFIRMED` once SOP had confirmed them. In some instances, the confirmation from SOP could not be mapped
to the approved ledger, which would result in the ledger remaining in an `APPROVED` state and a new `UNKNOWN CREDIT` ledger
being created and confirmed instead.

Ledger status is not a meaningful concept in the new system, as we are the source of truth for financial transactions. However,
this leads to a situation where balances may not reflect the actual debt position because of the erroneous `APPROVED` ledgers
that have been made duplicates through the process above.

## Decision

The solution is to omit `APPROVED` ledgers from queries that calculate the debt position, which now only consider `CONFIRMED`
ledgers. Additionally, the Invoices page will not display `APPROVED` ledgers, as they are duplicates of the `CONFIRMED` `UNKNOWN CREDIT` 
ledgers.

## Consequences

Omitting `APPROVED` ledgers from the invoice transaction list may potentially cause some confusion for users, as this is 
different from the existing Sirius finance tab. However, the existing approach is also confusing, as it does not accurately
reflect the debt position, where the transaction amounts come to a different total to the invoice balance. This will also
only affect legacy transactions, which users will know the source of truth for which is SOP.

The alternative to this was to create matching `APPROVED` ledger entries for zero the erroneous ledgers. However, this 
would still require the encoding of legacy processes in the new system, in the form of labelling to explain the purpose of
the ledgers, and would also require some way of excluding the new transactions from reports and journals.
