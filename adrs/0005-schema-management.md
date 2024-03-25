# 5. The Finance API service will takeover ownership of the `supervision_finance` schema

Date: 2024-03-14

## Status

Accepted

## Context

Currently, `opg-sirius` is responsible for the management of data in the Sirius database, including the `supervision_finance` 
schema. However, the Finance API will not only be the primary (and eventually _sole_) user of the data, but also the most
likely cause for changes to the schema definition (i.e. it is where active development on the tables is currently taking
place). This is a danger as it leads to two separate locations for the management of the DB as the single source of truth, 
which could lead to services deviating, deployment and replication issues, etc.

## Decision

Ownership of the `supervision_finance` schema will be transferred to this repository, and any future migrations to this 
schema will be recorded and actioned from scripts here. To reduce Sirius' exposure to the schema, we will look to condense
the existing Sirius migrations so there is a clean start in the main `opg-sirius` repository, meaning it will be easier
to spot if invalid migrations to the schema are attempted for any reason.

This is with the long term aim of removing Sirius's write access to the schema, potentially also removing read access also,
or even separating the schema to a different database.

## Consequences

Shared write access to a single data source by multiple services is considered a microservice anti-pattern and carries
with it significant risk if not managed correctly. However, we feel this risk is limited in this project due to the
small number of developers making changes, and being able to monitor and control changes while transitioning to a 
situation where Sirius does not need write access to the schema.

If we want to further mitigate these risks without rewriting sensitive business logic around invoice generation, we could
develop write endpoints on the Finance API for Sirius to call instead of persisting directly in the database.