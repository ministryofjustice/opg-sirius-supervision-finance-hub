# 28. Working days calculation

Date: 2025-08-14

## Status

Accepted

## Context

A direct debit collection should be made at least 14 working days after the instruction. As working days vary each year,
we need a mechanism to obtain these dates in order to be able to make the instruction.

## Decision

There are a number of libraries and APIs we could use but the simplest is to use the Gov.uk Bank Holidays API (https://www.gov.uk/bank-holidays.json).
This requires no authentication, serves simple JSON, and is reliably updated. We can improve performance by caching the data, 
and for that we can utilise our existing in-memory caching used for storing user data. 

## Consequences

The cache refreshes every 12 hours, so we shouldn't run into issues with stale data. Even if the cache did not refresh for some
reason, the API returns the next 2 years of holidays, and we are only ever looking at most one month ahead.

There is also the possibility that the API becomes unavailable. However, as a statically served JSON file, this is very
unlikely. Even so, logging has been put in place should this happen.
