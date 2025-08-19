# 29. Pending Ledgers

Date: 2025-08-18

## Status

Accepted

## Context

As per ADR-0015, ledgers are designed to be immutable, and we made changes to how pending invoice adjustments were stored
in order to maintain that. We now have a similar issue in regard to Direct Debit schedules, where collection of payment
takes place at a date in the future, and follows a presumptive model whereby we only receive confirmation of failed payments.

We have two options: 
* Recording the payment in a separate "pending" table and have a daily task to convert payments into 
ledgers based on collection date.
* Create the ledger in a "pending" state, without conflicting with the immutability constraint.

## Decision

We will create future collections as ledgers on a successful response from AllPay, using the collection date as the dates
for the ledger, effectively making it "pending" until the collection date has passed. The Invoices tab will then be amended
to display ledgers with future dates as pending. These will automatically change to be displayed as confirmed once that date
passes.

There are a few issues to resolve in this approach:

1. Our financial reports rely on these dates to filter the data but are based on the assumption that all dates will be in
   the past. These reports will need to be amended to ensure that only confirmed ledgers (i.e. have current or past dates).
2. The `created_at` date currently serves two purposes: It is the auditable field for record creation but also used as
   the confirmed/fulfilled payment date. This would need to be split out to maintain an auditable field.
3. Billing history would need to be amended to take account of future dates. The billing history is ordered by create
   date, and this would be used for a "Direct Debit payment instruction" event, but this event shouldn't affect the balance,
   despite being a ledger. There should then be a second event for the ledger affecting the balance on the collection date.

As a result, we will need to use a different date to `created_at` for this future date. There are currently three ledger 
dates in use:

* `ledger::datetime` - payment/upload date: The date reported as processed by the bank.
* `ledger::bankdate` - received/created/bank date: The date the payment file covers (could contain multiple of the above dates).
* `ledger::created_at` - auditable field/confirmed/fulfilled payment date: The date the ledger is created in the database.

We will create a new column to record this future date to not interfere with the audited field. We will set its value to
the `created_at` field on creation to duplicate across the values, as this will be the case for all records other than
AllPay Direct Debits.

## Consequences

All reports will need to be updated to use this newly created field. `created_at` was a field introduced in 2024, so many
historical records will not have a value to duplicate. However, this is not an issue as the reports that rely on it do not
report on pre-go-live data.
