# 40. Court reference updates in Finance Client

Date: 2026-05-27

## Status

Accepted

## Context

The original intention of enforcing writes to the `supervision_finance` schema to this service was well-intentioned, but
has led to some issues in practice. One is that the `finance_client`'s court reference is used as an identifier in both
Allpay and Payments, but if the `client_created` event is played out of turn, e.g. the event is received before the 
client is written to the database by Sirius, then all subsequent events fail.

## Decision

As the Finance Client is created by Sirius in the first place, the simplest solution is to write the court reference to
the entity at creation.

## Consequences

This doesn't weaken the separation of concerns, as the Finance Client was being created by Sirius anyway, but it doesn't
improve the situation. There were two other options considered. First was to write the whole Finance Client record in
response to the `client_created` event, which would guarantee the data was available, but due to the highly coupled nature
of Sirius and its test suite, this would have been a significant amount of work for little benefit. The second was to not
duplicate the court reference at all and instead fetch it from the `public` schema when required, but this would have 
meant every query joining on another table, and further weakening the separation of concerns.
