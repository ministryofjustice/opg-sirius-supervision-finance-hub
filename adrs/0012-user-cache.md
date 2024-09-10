# 12. In-memory cache for user data

Date: 2024-07-01

## Status

Accepted

## Context

In a microservice architecture, it is desirable for each service to own its data, which usually means a data source should
only have a single means of access. We have compromised on this principle with the `supervision_finance` schema in the 
initial phase of the Payments project as existing Sirius functionality will require read/write access to the tables until
the functionality is replaced. However, the Payments service now has a requirement on Sirius data, namely the `assignees`
table for user data in order to map billing history events with the users that generated them.

There are two options for accessing this data:
1. Provide the Payments service access to the `public` schema, so it can read from the `assignees` table directly
2. Fetch user data from Sirius via an API

## Decision

User data is now fetched from Sirius via the existing user list API (`api/v1/users`). As this user data changes infrequently,
this data is cached locally using `go-cache`, a well-used cache implementation. The API call is made by the front end
service for two reasons. First, from an architectural perspective, this enforces a separation of concerns, i.e. the back
end service is only responsible for its own data, and second, this allows us to piggyback off the user's session in the 
request to Sirius.

## Consequences

There are a number of trade-offs and potential consequences of this decision:
1. In-memory cache: As the cache is in-memory only, it will have to be fetched for each running service and will not be
   persisted when a service is redeployed. This would be an issue at scale but Sirius will only be running 1-3 instances 
   concurrently and refreshing the cache takes a second at most, so this is acceptable. We also have the option of moving
   to a persistent or distributed cache at a later date if this does cause issues.
2. Security: The user list API is available to all Sirius roles and returns a full list of users, including ids and emails.
   This is not great from a security perspective, and while it is out of scope of this task to improve security, it could
   impact this solution if this is addressed separately and access is restricted, as we may potentially no longer have
   access to the users. Although Sirius and Payments are currently managed by the same team so this can be managed
   appropriately, this isn't guaranteed in the future, so we need to be aware of this going forward. Contract testing
   would mitigate this.