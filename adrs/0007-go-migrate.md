# 7. Adopt go-migrate for data migration

Date: 2024-03-19

## Status

Superseded by ADR 0009-goose

## Context

As per ADR#0005, this project will be the owner of the `supervision_finance` schema going forward. As a result, we need
a mechanism by which to generate and run migrations against the database, both locally and in production. This needs to 
be a simple tool to use, allow for rollbacks and be able to be applied in multiple different contexts (e.g. local/dev/production).

## Decision

We will use the go-migrate tool for generating migrations. This creates separate SQL-only up and down migrations in numerical
order that can be be applied and rolled back, and is also compatible with sqlc for generating the data layer of `finance-api`.

## Consequences

go-migrate is a widely-used and well-maintained tool, so we do not anticipate there to be any negatives to adopting this tool.
