# 26. Switching Allpay mocking to Imposter

Date: 2025-09-05

## Status

Accepted

## Context

We are currently mocking the Allpay API using Prism and an OpenAPI spec (ADR 00026). This serves its purpose for local 
development but has limited flexibility when it comes to conditional responses. In addition, Allpay have confirmed the 
functionality of the development environment we have access to, and while it does allow us to access the real API for testing
and demos, there is no way to clear data or set up fixed responses, as the data is automatically cleared daily and there
is no ability to add fixtures.

## Decision

[OPG Paper Identity](https://github.com/ministryofjustice/opg-paper-identity) have faced similar challenges with integrating
with 3rd party APIs and also use OpenAPI to document the service. The project adds dynamic functionality on top by using
[Imposter Mocks](https://docs.imposter.sh/), which allow for fixed responses that are conditional on the attributes of the
request (e.g. path or query parameters, request body variables, etc.). This not only means we can get rid of the prefer 
header (which required passing around cookies in production code for test purposes), it also allows this mocking behaviour
to be used in live non-prod environments. As the conditional responses are based on properties of the request rather than
headers or cookies (e.g. client surname), these conditions can be triggered via the UI, allowing all users to test and 
demo behaviours.

## Consequences

None.
