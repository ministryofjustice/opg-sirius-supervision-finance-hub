# OPG SIRIUS SUPERVISION FINANCE HUB

### Major dependencies

- [Go](https://golang.org/) (>= 1.22)
- [docker compose](https://docs.docker.com/compose/install/) (>= 2.0.0)

#### Installing dependencies locally:
(This is only necessary if running without docker)

- `yarn install`
- `go mod download`
---

## Local development

The application ran through Docker can be accessed on `localhost:8888/clients/1/invoices`.

To enable debugging and hot-reloading of Go files:

`make up`

Hot-reloading is managed independently for both apps and should happen seemlessly. Hot-reloading for web assets (JS, CSS, etc.)
is also provided via a Yarn watch command.

Both the `finance-hub` (front end) and `finance-api` (back end) can be debugged independently, as they expose different
ports for Delve:

* `finance-hub`: 2345
* `finance-api`: 3456

-------------------------------------------------------------------
## Run the unit/functional tests

`make unit-test`

-------------------------------------------------------------------
## Run Trivy scanning

`make scan`

-------------------------------------------------------------------
## Architectural Decision Records

The major decisions made on this project are documented as ADRs in `/adrs`. The process for contributing to these is documented
in the first ADR.