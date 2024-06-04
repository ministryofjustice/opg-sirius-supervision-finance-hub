# OPG SIRIUS SUPERVISION FINANCE HUB

### Major dependencies

- [Go](https://golang.org/) (>= 1.22)
- [docker compose](https://docs.docker.com/compose/install/) (>= 2.26.0)
- [sqlc](https://github.com/sqlc-dev/sqlc?tab=readme-ov-file) (>=1.25.0)
- [goose](https://github.com/pressly/goose) (3.19.0)

#### Installing dependencies locally:
(This is only necessary if running without docker)

- `yarn install`
- `go mod download`
---

## Local development

The application ran through Docker can be accessed on `localhost:8888/finance/clients/1/invoices`.

To enable debugging and hot-reloading of Go files:

`make up`

Hot-reloading is managed independently for both apps and should happen seamlessly. Hot-reloading for web assets (JS, CSS, etc.)
is also provided via a Yarn watch command.

Both the `finance-hub` (front end) and `finance-api` (back end) can be debugged independently, as they expose different
ports for Delve:

* `finance-hub`: 2345
* `finance-api`: 3456

-----
## Generating the sqlc Store

The data access layer of `finance-api` is auto-generated with `sqlc`. It reads the database schema (specified in `/migrations/`)
to generate the models and queries in `/finance-api/internal/store/queries/`.

To generate these files after making changes, run `make sqlc-gen`.

## Generating migrations

To generate migration files with `goose`, install it locally with `brew install goose` and run:

`goose -dir ./migrations create <name-of-migration> sql`

Or copy the up and down files and increment them.

## Adding seed data

In general, look to keep tests self-contained by having them create (and clear) the data they require for testing. However,
there are times where we need to seed the data in advance in order to test it, such as where the method for adding the 
data is driven by Sirius, e.g. Cypress tests asserting on the client header.

To seed this data, add the inserts to `/test-data.sql`.

-----
## Run the unit/integration tests

`make test`

## Run the Cypress tests

`make cypress`

## Run Trivy scanning

`make scan`

-----
## Architectural Decision Records

The major decisions made on this project are documented as ADRs in `/adrs`. The process for contributing to these is documented
in the first ADR.

-----
## HTMX & JS

This project uses [HTMX](https://htmx.org/) to render partial HTML instead of reloading the whole page on each request. However, this can 
mean that event listeners added on page load may fail to register/get deregistered when a partial is loaded. To avoid this,
you can force event listeners to register on every HTMX load event by putting them within the `htmx.onLoad` function.

HTMX also includes a range of utility functions that can be used in place of more unwieldy native DOM functions.
