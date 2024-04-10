# 9. Switch to Goose for data migration

Date: 2024-04-09

## Status

Accepted

## Context

During the process of setting up the infrastructure and deployment pipeline for the finance services, it was discovered
that go-migrate was not as well-maintained as it previously appeared, with multiple vulnerabilities in both the package
and the official Docker image, and their Github showed little progress in patching them.

## Decision

We have switched to [Goose](https://github.com/pressly/goose), which is the second most widely used Golang migration tool.

## Consequences

Goose does still have one vulnerability at present, although it has already been patched and is due for release this month.
Releases have consistently been monthly, which gives confidence going forward.
