# 31. Cancelling Direct Debit Mandates

Date: 2025-10-06

## Status

Accepted

## Context

BACS can take up to three working days to take a payment for a Direct Debit schedule instruction. This means that when cancelling
a mandate with pending collections, there is a period where the collection may collect even if the mandate was closed.

## Decision

To avoid this behaviour and risk our records and Allpay's becoming out of sync, we will choose a closure date that accounts
for this. If a pending collection is within three working days, we will take the last pending collection within that group
and set the closure date for the next working day. This ensures all pending collections that may already be in progress are
collected. The remaining pending collections will have their schedules removed in Allpay and are set to cancelled in our 
database.

## Consequences

This may mean more schedules are collected than would have been had we closed the mandate immediately. However, we have a
refund process in place for this, and it is a better option than risking our sources of truth from deviating.
