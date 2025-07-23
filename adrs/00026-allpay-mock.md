# 26. AllPay Mock

Date: 2025-07-08

## Status

Accepted

## Context

We are using AllPay as our Direct Debit bureau and will be using their API for creating and closing mandates, billing
customers with payment schedules, and fetching failed payments. 

In order to assist development against the API, we need to create a mock service. We may also need to be able to deploy 
the mock in development environments where we aren't wanting to hit the actual API.

## Decision

As we have experience with json-server, this was the initial preference. However, this is complicated by the bearer token
required for authentication, the scheme code, and the base64 encoded client details as path parameters. We also don't
need to test against client-specific data, and fixed responses per route and response type will suffice.

Additionally, it will be useful to have a defined API specification that we can compare against the documentation and update
as needed. Therefore, we have converted the documentation to an OpenAPI spec and will use Prism to run a mock from it.
This is deployable as a Docker container with no additional configuration.

## Consequences

Prism does not allow for much in the way of conditional responses. We can get different responses by passing in
`prefer` headers, which is how we handle this requirement in json-server, but if we do need it to be more dynamic in how
this is handled, we would need to implement the mock in a custom application instead.
