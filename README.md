# OPG SIRIUS SUPERVISION FINANCE HUB

### Major dependencies
- [Go](https://golang.org/) (>= 1.22)
- [docker compose](https://docs.docker.com/compose/install/) (>= 2.26.0)
- [sqlc](https://github.com/sqlc-dev/sqlc?tab=readme-ov-file) (>=1.25.0)
- [goose](https://github.com/pressly/goose) (3.20.0)
- [htmx](https://htmx.org/) (2.0.0)
- [pgx](https://github.com/jackc/pgx) (5.5.5)
- [validator](https://github.com/go-playground/validator) (10.19.0)

### Walkthrough
A [walkthrough](docs/walkthrough.md) of this project has been written for the Golang Community of Practice, describing 
the package structure and some technical aspects of the codebase.

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
there are instances where we need to seed the data in advance in order to test it, such as where the method for adding the 
data is driven by Sirius, e.g. Cypress tests asserting on the client header.

To seed this data, add the inserts to `/test-data.sql`.

-----
## Run the unit/integration tests
`make test`

## Run the Cypress tests
`make build-all`
`make cypress`

Or to run interactively:

```
cd cypress
npx cypress open baseUrl=http://localhost:8888/finance
```

## Run Trivy scanning
`make scan`

## Run adhoc tasks
To create an adhoc task to be against the service in a live environment:
* Validate the task name in `service/process_adhoc_event.go`. If more than one adhoc task is needed, use this service to
  conditionally handle the events by task name.
* Add the logic to be performed in `/api/process_adhoc_event.go`.

Then, in the AWS environment you wish to run the task:
* Go to Amazon EventBridge -> Event buses
* Select the  `<env>-supervision` event bus
* Click "Send events"
* Enter the "Event source" as `opg.supervision.finance.adhoc` and the "Detail type" as `finance-adhoc`
* Enter the expected JSON into "Event detail" i.e. `{"task":"<task-name>"}`
* Click "Send"

-----
## Architectural Decision Records
The major decisions made on this project are documented as ADRs in `/adrs`. The process for contributing to these is documented
in the first ADR.

-----
## HTMX & JS
This project uses [HTMX](https://htmx.org/) to render partial HTML instead of reloading the whole page on each request. 
However, this can mean that event listeners added on page load may fail to register/get deregistered when a partial is 
loaded. To avoid this, you can force event listeners to register on every HTMX load event by putting them within the 
`htmx.onLoad` function.

HTMX also includes a range of utility functions that can be used in place of more unwieldy native DOM functions.

-----
## Mock APIs
This service integrates with multiple APIs, and we have a number of different approaches to mock responses.

### Sirius
Sirius endpoints are mocked using [json-server](https://github.com/typicode/json-server). This is a simple Express app that
reads responses from a JSON file. The config files are located in `/json-server`, with routes specified in `routes.json`
and the data in `db.json`. Additional middleware can be written in JS to intercept requests.

### Allpay
The Allpay Direct Debit API is mocked using [imposter](https://docs.imposter.sh/). This applies a config to an OpenAPI spec
and responds with a file or string based on request data. Config files are located in `/api-mocks/allpay`.

Note that Allpay requires client reference and surname path parameters in the URL to be base64 encoded, as these strings 
may include invalid characters (e.g. whitespace, reserved symbols). When the mock switches response based on a path parameter,
the value will need to be the base64 encoded string, with the actual value included as a comment in the line above.

### Bank Holiday API
We use the GovUK bank holidays endpoint in order to calculate working days. This is just a JSON file, but to avoid calling
this repeatedly during testing, there is a simple Go file to serve it instead, located in `/api-mocks/holidays-api`. The
`bank-holidays.json` file contains all dates from 2024-2027, and this can be manually updates as new dates are available.
