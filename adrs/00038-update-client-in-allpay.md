# 38. Update Client in Allpay

Date: 2026-05-18

## Status

Accepted

## Context

Allpay uses a combination of court reference and surname as identifiers within the URL for API requests. As a result, we
need a mechanism to update the surname in Allpay, otherwise there will be a mismatch between the client's surname in Sirius
and Allpay.

## Decision

There is an existing `client-created` event that is triggered when a client is created and edited in Sirius, but the event
was redundant (and badly named) for the edit route, as it would only trigger for court reference changes, and the court
reference cannot be edited. This event has been replaced with a new `client-updated` event that is triggered when the client's
surname is changed, and includes both old and new values in the payload.

## Consequences

In order to update the surname in Allpay, the client's address must also be included. This is only available in the `public`
schema. The original intention of the Payments project was to keep the schemas entirely separate and only interact with the
`supervision_finance` schema, which is why the court reference was duplicated onto the `finance_client`. However, this has 
already been relaxed in previous changes to get the surname, so we will read from the public schema to fetch this data. 
This is not ideal from a clean architecture perspective, but is the pragmatic solution and there is no hit to data integrity
as there is still no write access to `public` from Payments.

What this does mean is we can probably simplify some of the existing functionality where we fetch client data from Sirius
in the UI, whereas we could now fetch it direct from the database.

The event has been architected to allow for additional changes to be included, although that is not currently required.
