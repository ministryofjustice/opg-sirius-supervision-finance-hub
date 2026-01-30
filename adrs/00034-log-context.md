# 34. Contextual logging

Date: 2026-01-27

## Status

Accepted

## Context

Logging in Payments services includes useful data but are not categorised in a way that makes it possible to identify and
filter logs based on where they originate. This is needed in order to be able to create log-based metrics and alerts for specific
areas of functionality, such as the Allpay integration.

## Decision

A "category" field will be added to the log structure, and all logs will be assigned a category based on their origin. We 
can do this by adding a logger function at the component level for each service. The categories will be defined as follows:
- `allpay` - Allpay integration.
- `api` - API requests and responses in the backend API service.
- `application` - Application logic and business processes in the backend API service.
- `auth` - Authentication and authorization processes.
- `handler` - Request logs in the frontend service.

## Consequences

This change allows us to filter and analyse logs based on their categories, making it easier to monitor specific areas of 
functionality. It also brings us closer to the structure of Sirius logs. However, the log structure is not fully standardised,
as the Sirius logs are more nested and include additional information we have not included, e.g. user roles. Further standardisation
may be considered in the future.
