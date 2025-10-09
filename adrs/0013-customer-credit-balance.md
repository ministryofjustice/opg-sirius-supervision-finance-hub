# 13. Customer Credit Balance

Date: 2024-07-23

## Status

Accepted

## Context

Managing client payments not only means allocating funds to debits on invoices but handling any overpayments or adjustments
that may be made in error or due to changing situations. This forms credit on the account, known as the "Customer Credit
Balance" (CCB). At any time, we need to know the amount of credit on account, and be able to reapply those funds to new invoices.

There are two ways credit can be added on account:

* *Unapply* is when a discount is applied to already cleared debits (e.g. a fee reduction has been successfully applied 
  for after an invoice has had the full amount paid). A new ledger would be created with one allocation for
  the fee reduction being credited on the invoice and an allocation for the same amount being debited against the invoice.
* *Overpay* is when a payment is larger than the existing debt on the account. A new ledger would be created with an 
  allocation for each invoice being paid, and a separate allocation for the remainder, not associated with any invoice.

And two ways credit can be removed:

* *Reapply* is where a new invoice or debt is created and the CCB is partially or fully applied. A new ledger would be created
  with an allocation for the amount being applied to the invoice.
* *Refund* is where the CCB is returned directly to the fee payer. This process has not been defined as it is dependent
  on our payment provider integration, but it is expected a new ledger would be created with an allocation for the full 
  amount, not associated with any invoice.

## Decision

Instead of recording Customer Credit Balance as a discrete entity, we have decided we can use the existing schema of 
ledgers and allocations to record the movement of credit. This can be done with two new `ledger_allocation` statuses:

* `UNAPPLIED`
* `REAPPLIED`

This can then be used in combination with the `invoice_id` column to categorise the transaction type:

* Unapply = `status = 'UNAPPLIED' AND invoice_id IS NOT NULL`
* Overpay = `status = 'UNAPPLIED' AND invoice_id IS NULL`
* Reapply = `status = 'REAPPLIED' AND invoice_id IS NOT NULL`
* Refund  = `status = 'REAPPLIED' AND invoice_id IS NULL`

It then makes it simple to calculate the CCB at any point in time, as you simply subtract the total ledger allocations with 
a status of `REAPPLIED` from the total with a status of `UNAPPLIED`.

## Consequences

As this is adding new values to an existing field, we need to be wary of two things:

* Existing implementations (i.e. Sirius) need to be able to handle the additional values. This shouldn't be an issue going
  forward as Sirius shouldn't need to perform any queries on ledger allocations, but it should still be considered and 
  investigated.
* Confusing the semantics of the `status` column. The original plan was for the allocation type to store the new values,
  but on inspecting the schema, allocations don't have a type themselves and instead are the type of their parent ledger.
  That said, the `status` column currently contains the following statuses: `PENDING`, `ALLOCATED`, and `UN ALLOCATED`, so
  these new statuses don't feel to change the meaning of the column drastically.