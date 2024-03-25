# 4. Finance front ends and back ends will be developed within the same monorepo 

Date: 2024-03-14

## Status

Accepted

## Context

All new services since breaking out of the opg-sirius monorepo have been in their own repositories. However, in keeping 
the front end and back end services in the same repo, there are potential benefits:

* Having both services share the same DTO structs that encode JSON data provide a form of contract between them
* Easier to reason about two services that exclusively serve each other

However, there was uncertainty about how to do this in practice with Golang modules and how CI/CD would work with 
independent but coupled services. 

Proof of concept to prove multiple independent Go services can be built and deployed using a single go.mod file, while 
each sharing code and data that falls outside their package.

This POC has the following architecture:

Top level go.mod file, defining the dependency structure of the module
A user-facing service - "finance-hub" - that serves the UI
A back-end service - "finance-api" - that provides the data access and business logic for user requests
A shared package - "shared" - that contains the JSON structs that are used to transfer data between the services
To prove this worked, I did the following:

The existing /users/current route was moved so it is served by the finance-api service
The finance-api service reads from a postgres database and returns the value using the shared data structs
All services and dev dependencies have been dockerised
Both services have hot-reloading when run in dev mode and independent debugging is available (hub is on port 2345 and api is on port 5432)

## Decision

Following a Proof of Concept, we made the decision to adopt the monorepo approach with the following packages:

* A user-facing service - "finance-hub" - that serves the UI
* A back-end service - "finance-api" - that provides the data access and business logic for user requests
* A shared package - "shared" - that contains the JSON structs that are used to transfer data between the services

## Consequences

This should make it easier to reason about "Finance" as a distinct workstream within Supervision, which may have further
positive effects on future architectural decisions. It also provides some level of contract testing for the API without
having to introduce additional tools or processes. 

The downside is this does tightly couple the two services. However, the intention is for the front end to exclusively 
communicate with the back end, and the back end to exclusively provide an API for the front end, so they are tightly
coupled by definition. Further, this is only ever going to be a project worked on by a small number of developers on the 
same team, so the dangers of tight coupling are diminished.