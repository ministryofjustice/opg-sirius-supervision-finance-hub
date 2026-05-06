# 37. Schedule Removal

Date: 2026-05-06

## Status

Accepted

## Context

During the transition away from the previous provider to Allpay, Billing created mandates with future schedules as placeholders,
with the intention that these would be overwritten when the actual schedule was set up. This was done due to poor communication
resulting in us not knowing about pre-schedule mandates, which would have rendered this unnecessary, and the fact that we
are not using the PUT schedule API, as we don't want to overwrite existing schedules during the normal course of business.

As a result, we need to clear up around 12,000 schedules that are due to be collected for debt that does not exist.

## Decision

We will use our existing admin file upload functionality to upload a CSV of schedules to be removed. This will be sent from
Finance Admin to Finance API, which will need to call the Allpay API to remove the schedules. As this will be many thousands
of schedules at a time, we will need to do this asynchronously. We could do this in batches, but we have no guarantees from
Allpay for response times, so we will instead dispatch an event for each schedule to be removed, which will then be sent
back via Eventbridge for each schedule to be removed asynchronously.

## Consequences

This is a similar approach to how we have triggered and replayed schedule creation, so it is an established pattern. However,
unlike the other uploads, we won't be able to send a success/fail email to the user as the schedules will be removed in 
a separate async process. As this is likely to be a one-off task, we can monitor the process in the logs and DLQ for the
events, but if this were to be a permanent feature, we may want to consider creating user tasks on the client for failed
removals, as the Allpay API documentation does state that removing schedules can fail if the schedule cannot be clearly 
identified, and that removal will need to be retried manually.
