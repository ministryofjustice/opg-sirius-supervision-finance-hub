# 14. External Events

Date: 2024-09-10

## Status

Accepted

## Context

Although payments and invoice management will be managed in this new service after go-live, Sirius will still be generating
the majority of invoices, or at the very least, producing the events that trigger invoice creation, e.g. order made active,
death of a client, etc. As excess customer credit balance can be reapplied to new invoices on creation, this service needs
a way to react to these events.

## Decision

We have chosen to adapt and utilise the existing EventBridge pattern for event-driven messaging between services. The
`InvoiceCreated` event handler in Sirius sends a `debt-position-changed` event to the Supervision event bus in AWS EventBridge.
This event is then forwarded to the `/events` endpoint in the Finance API, where it is handled. This not only allows us to
react to external events, but do so asynchronously. In the future, we also have the option of using this architecture as
an event bus for internal events we want to handle in an event-driven, asynchronous manner.

## Consequences

The consequences of this change are what to expect from adopting an event-driven microservice architecture. It is harder 
to test, as not only are events coming from an external service, they are routed through a vendor infrastructure we have 
no control over. As EventBridge is backed by AWS SQS, which is "at-least-once" delivery, it is possible the same event 
could be received more than once. As a result we need to be cautious to ensure event handlers are idempotent (which reapplied 
already is). Additionally, there are potential issues to be expected with any asynchronous process, and we will have to 
consider how we handle those in both future architecture decisions and UX.