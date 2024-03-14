# 2. Create a new Finance back end to service the Finance Hub

Date: 2024-03-14

## Status

Accepted

## Context

The Finance Hub requires an API to fetch finance-related information and perform payment-related actions. The existing Sirius API exists, however we would like to not use it for the following reasons:

* The API would need to be versioned to avoid new behaviour or data shapes impacting existing processes
* Business logic would also need to be versioned if behaviour changed
* There is a desire long-term for the Sirius monolith to be smaller

Additionally, there are potential benefits from a new, independent API:

* Will eventually allow us to truly separate the finance schema from the rest of Sirius (compliance/regulatory benefits?)
* Development in Golang over PHP
* Contract testing will be easier than retrofitting to Sirius, and some of it may come for free by sharing the data structures between FE and BE (e.g. in a separate shared module)
* Likely to be our best opportunity to stake a stab at splitting up Sirius

## Decision

We will create a new back end service to for the Finance Hub. This will provide the following:

* Data access for data to be viewed in Finance Hub (e.g. invoices)
* Any new/amended functionality for finance-related actions (e.g. making a fee reduction)
* Interface with the new payments provider

To reduce the scope and risk, it will not touch the invoice generation process in its initial iteration.

## Consequences

This will hopefully allow us to develop the service with less risk of affecting existing processes, as this is a new 
service that can be switched to, rather than having to version or feature-flag. This should also make it easier to test
in isolation, and builds towards our longer-term aims of smaller, more focused services.

The risks are that we don't currently have a pattern for purely back end Golang microservices, so this will likely
decrease developer throughput in the short term. There are also questions regarding authentication and data sources that
will require further investigation.
