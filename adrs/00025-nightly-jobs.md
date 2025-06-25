# 25. Nightly jobs

Date: 2025-06-19

## Status

Accepted

## Context

Refunds that have not been fulfilled need to expire after two weeks of their last status update, as these contain bank
details that should only be kept for a short period. To do this, we need a mechanism for triggering an event that will
find and expire refunds on a scheduled (e.g. nightly) basis.

## Decision

In Sirius, we use CloudWatch events to deploy tasks in ECS for particular jobs based on cron expressions. As we have 
an established process for using EventBridge for passing messages between services, we can use this instead. Instead of
deploying a task, the CloudWatch rule will target the Supervision event bus in the same way events created by Sirius or 
Payments do. This not only simplifies the infrastructure, it also means that authentication and system user management
is already handled, as the passthrough lambda that forwards the events already encodes the API key as the bearer token.

## Consequences

As with all SQS messaging (which EventBridge is built on), it is "at-least-once" delivery, so the same event could be 
processed more than once. This isn't an issue in this case as rejecting or cancelling a refund is idempotent, but it is
an important consideration if we use the same process for future events.
