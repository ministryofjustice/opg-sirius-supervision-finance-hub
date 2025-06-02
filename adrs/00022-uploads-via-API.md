# 22. Uploads via API

Date: 2025-05-19

## Status

Accepted

## Context

As detailed in ADR-0016, the Finance Admin back end was deprecated and partly absorbed into the Finance API. However,
file uploads was retained due to there being no pressing need to move it. It does however require a more complex process
than is necessary, due to requiring EventBridge and S3 to pass the files to the Finance API.

## Decision

The following changes will be made:
* Send upload files directly to Finance API
  * This requires changing Finance Admin to send as an API request
  * And a change in Finance API to move endpoint from `/events` to a new top level API
  * Finance API will process file in a goroutine async
* Remove finance-admin-api package
* Remove build steps
* Remove unused infrastructure

## Consequences

The process will be significantly simpler and cheaper to run, as there will be no S3 storage, EventBridge messaging, or
additional back end container.