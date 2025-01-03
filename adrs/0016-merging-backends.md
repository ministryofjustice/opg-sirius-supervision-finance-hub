# 15. Merging Backends

Date: 2024-12-20

## Status

Accepted

## Context

The initial plan for the Finance Admin service was that it would eventually encompass other Supervision reporting functionality.
This, along with a desire to work on admin functions and payments logic in parallel, led to the decision to implement the
Finance Admin backend as a separate service to the existing Finance Hub backend. While this had its advantages, such as
a separation of concerns and a read-only database connection for reporting, it also led to a number of issues:
* A more complex architecture, requiring events and storing data in S3 to be sent between the services to trigger operations
* A lack of ownership of data on the Finance Admin side, as it was neither in control of the schema or migrations, or the
  methods of data creation
* Difficulty in testing, due to the above

## Decision

We have decided the best course of action is to merge the Finance Admin backend into the existing Finance Hub backend. This
involved the following steps:
* Adding the required endpoints to the Finance Hub API
* Adding additional features to the filestorage client
* Adding the Notify client
* Refactoring migrated code to take account of the different code architectures
* Adding a new read-only database connection to the Finance Hub backend
* Rewriting of migrated report tests
* Removing surplus architecture

## Consequences

In merging the services, the architecture is simpler, with fewer parts to maintain and fewer points of failure. Although
this does mean that the Finance Hub backend is now responsible for more functionality, its scope as a finance service hasn't
changed. This change does mean that the Admin service cannot now easily be expanded to be a wider Supervision reporting 
service, this has been argued against anyway and is no longer a future requirement.