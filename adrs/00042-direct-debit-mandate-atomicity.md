# 42. Direct Debit mandate atomicity refactor

Date: 2026-07-08

## Status

Accepted

## Context

When creating a Direct Debit mandate for a client with outstanding debt, the system first calls Allpay to create the mandate,
then calls Allpay again to create a collection schedule for the outstanding balance. This introduces a failure mode where 
the mandate creation succeeds but the schedule creation fails, leaving the client with an orphaned mandate and no pending 
collections. However, unlike other API failures where the user gets an error message and is able to retry, the client's
payment method is updated as expected and the retry would fail because the mandate already exists.

## Decision

Considering all the Allpay interactions, creating a schedule is the only operation that is neither user-driven nor event-driven.
Event-driven interactions already have a retry and triage mechanism, thanks to our Eventbridge implementation, while user-driven 
interactions can be retried by the user and any subsequent failures can be flagged for manual intervention. Therefore, we
could either implement a retry mechanism for schedule creation or refactor the code to use Allpay's mandate API, as opposed
to their pre-schedule mandate API, which allows for the creation of a mandate and schedule in a single call.

In doing so, we have also refactored the code to centralize the schedule creation logic, which was previously duplicated in two places.

## Consequences

This removes the failure mode that leaves orphaned mandates and allows users to properly retry failed requests. It also
simplifies the codebase by centralizing the schedule creation logic, making it easier to maintain and reducing the risk of future bugs.

An alternative considered was to create schedules via the event-driven mechanism, e.g. by sending a message to Eventbridge
to loop back to the service with the schedule creation request. This would make it asynchronous and provide a retry and triage
mechanism, but isn't necessary, thanks to being able to create the mandate and schedule in a single call. However, this is
an option we could consider in the future if we encounter further issues with API availability.
