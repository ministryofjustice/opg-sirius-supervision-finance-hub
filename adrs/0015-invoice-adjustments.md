# 14. Invoice adjustments

Date: 2024-09-11

## Status

Accepted

## Context

In Sirius, pending invoice adjustments are created as a ledger and allocation with a `PENDING` status, and approving an 
adjustment updates the status to `ALLOCATED`. This causes a number of issues:
* Allocations are mutable, meaning they cannot be relied on as a true source of truth, which makes it hard to derive a
  billing history.
* The unapply/reapply process is triggered by allocation creation, whereas these adjustments are only allocated when approved.
* It doesn't make much sense semantically to have allocations that are not "allocated", or "unallocated" (rejected) allocations 
  that never exist in the first place - we want allocations to be an immutable record of the transactions on an account.

## Decision

To resolve these issues, a new `invoice-adjustment` table is created to track invoice adjustments. Pending adjustments 
create an entry in the table and only when the adjustment is approved do we create the ledger and allocations. This allows
us to use the same ledger creation and unapply/reapply logic we use for everything else.

## Consequences

Adjustments created using the existing logic will still exist. However, the go-live runbook will ensure that all pending
adjustments have decisions made and there is no requirement to display legacy adjustments in the Pending Invoice Adjustments
tab.

This will mean that many of the statuses will become redundant. Ledgers will only ever be created with a `CONFIRMED`/`APPROVED`
status (tbc) and so we could deprecate the field, while `PENDING` and `UN ALLOCATED` allocation statuses would also be unused,
so a future data cleansing could tidy up and rationalise this.
