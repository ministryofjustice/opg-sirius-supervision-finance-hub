# 20. Payment reversals

Date: 2025-05-02

## Status

Accepted

## Context

We needed a way of changing data within the finance system in a programmatic way and not just running sql straight in the database.

## Decision

It was decided to go with an event approach through AWS, so that we can use all the logic in the finance system and because it is flexible for us to change when needed.
Another reason we went in this direction was so that we did not have to change any of the infrastructure when running the adhoc event.

## Consequences

As a result of this approach, we need to make sure that the event that is being triggered is safe to be run many times. As there could be a time
when the event fails the first time and tries to send the event multiple times.
