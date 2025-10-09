# 29. Future Collections

Date: 2025-08-20

## Status

Accepted

## Context

A direct debit collection should be made at least 14 working days after the instruction. This means we will need to record
the pending transaction in some way and only confirm the ledger once the collection date has passed. As ledgers are designed
to be immutable, we have two options:

1. Create the ledger with a future date
2. Record pending transactions in a separate table, then create the ledger on the collection date with a scheduled job

## Decision

We will go with option 2, and record the pending transaction in a separate table. A nightly task will be scheduled via 
EventBridge to find the transactions for the collection date and create the ledgers. Although option 1 is a cleaner 
approach and maps better to user expectations (e.g. seeing pending collections against invoices in the UI), this has some 
major advantages over the other approach:

* Recording ledgers for a future date would cause issues regarding unapplies and credit balance, as these would either need 
  to also be created with a future date, or subsequent transaction events (e.g. fee reductions) would need to take pending 
  transactions into account. Transactions created on the collection date would not have this issue as they would be treated
  as any other payment.
* Ledgers with future dates would create undefined behaviour in some of our reporting. For example, should a future ledger
  be accounted for in the Fee Chase report? Treating pending transactions as payments on the collection date instead mitigates
  these issues.

To implement this, we will create a new table to record the amount, client, and collection date. We will then use the same 
scheduled EventBridge process for expired refunds to find the pending transactions and create the ledgers.

## Consequences

Creating the ledgers with a nightly job does simplify the process, as the transactions can be treated as regular payments,
so no changes are required to any reports or other functions. 

However, payments differ from "pending" transactions in that a payment is applied to the oldest debt, while invoice adjustments
apply to the invoice they are applied to. This is why we are able to display the pending adjustment against the invoice.
Treating a future transaction as a payment means we are unable to do that, so we will be unable to display the pending 
transaction in the Invoices tab.

Having discussed this with Product, we feel this is more of a training/implementation issue, in that the process has not
changed from a user perspective. The only real change is that the schedule instruction is being sent by us rather than
SOP or Allpay. So while the user won't have much insight into what collections are due to be made, this isn't actually a
change in process. If this is required based on user feedback, we could add a Payments tab to display this information.
