# 44. Locking transactions

Date: 2026-08-17

## Status

Accepted

## Context

We encountered an issue where a user approving multiple adjustments in quick succession could cause a race condition where
an unapply is reapplied to an invoice that is being processed concurrently in a separate transaction. If this second transaction
has already fetched the invoice balance, it won't have seen the reapply and the invoice will be left with a negative balance.

This is because there is no locking on transactions.

## Decision

We will lock the finance client row by client ID when creating the transaction. As each transaction needs to fetch a lock
on the finance client row, this will prevent concurrent transactions for the same client, even if they are in different
processes and performing different operations.

In order to do this for batch upload processes (e.g. payments), we need to refactor the processing, as currently the whole
file is processed in a single transaction. In order to still account for the possibility of multiple lines in a file affecting
the same client, we will group the lines by client and process each group in a separate transaction. Some additional
failure handling is required, as an error in one line no longer aborts the whole file, but we still need to be able to
report back to the user which lines failed.

## Consequences

In order to simplify the transaction handling and ensure the lock is acquired, the transaction creation function assumes
the transaction is client-scoped. This is always the case in this service to date, but if transaction locking is required
for any other use case, this will need to be refactored.

We also need to be aware that locking rows in a transaction can cause deadlocks or degrade performance if the locks are
held too long. Therefore, we need to ensure that the transaction is only held for as long as necessary, such as by
performing validation and database reads prior to opening the transaction, and only acquiring the lock when we are ready
to write to the database.
