# 27. Fetching client data for creating a Direct Debit Mandate

Date: 2025-07-24

## Status

Accepted

## Context

The AllPay API for creating a Direct Debit Mandate requires client data that the Payments service does not have access to,
as it belongs to the Sirius `public` schema. In addition, we also need to validate against order and deputy information,
which suffers from the same problem. We have three options to obtain this information:

1. Cache all data ahead of time in the `supervision_finance` schema using events, in the same way we store the court reference.
2. Fetch the data asynchronously (e.g. event-driven) at form submission.
3. Call the Sirius API from `finance-hub` using the user's Sirius session.

## Decision

We will call the Sirius API for this data using the user's session token. The path `/supervision-api/v1/clients/{id}`
contains all the client information required (name and address), as well as data we can use as flags to validate an active
order and fee paying deputy. We already call this endpoint in `finance-hub` to get data for the client banner, so it would 
just be a matter of adding additional data expectations to the response.

## Consequences

Fetching the data in `finance-hub` means the front end service will be responsible for validation, whereas this is usually
done in `finance-api`. This should not cause issues but is a departure from established practice.

The clients endpoint in the Sirius API is very bulky and contains a lot of data that isn't relevant, e.g. every task ID
associated with the client. This is likely because it is used by multiple external services, as well as being used as the
internal API for the Sirius UI. This conflict could make the API very brittle, as a change for one service could result 
in other services breaking. As a result, we should introduce contract testing with Pact to ensure the API we are consuming
adheres to its contract.
