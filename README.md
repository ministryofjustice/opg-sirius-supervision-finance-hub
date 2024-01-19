# OPG SIRIUS SUPERVISION FINANCE HUB

### Major dependencies

- [Go](https://golang.org/) (>= 1.19)
- [docker compose](https://docs.docker.com/compose/install/) (>= 2.0.0)

#### Installing dependencies locally:
(This is only necessary if dunning without docker)

- `yarn install`
- `go mod download`
---

## Local development

The application ran through Docker can be accessed on `localhost:8888/supervision/finance/`.

**Note: Sirius is required to be running in order to authenticate. However, it also runs its own version of Finance on port `8080`.
Ensure that after logging in, you redirect back to the correct port (`8888`)**

To enable debugging and hot-reloading of Go files:

`make dev-up`

### Without docker

Alternatively to set it up not using Docker use below. This hosts it on `localhost:1234`

- `yarn install && yarn build `
- `go build main.go `
- `./main `

-------------------------------------------------------------------
## Run the unit/functional tests

`make unit-test`

-------------------------------------------------------------------
## Run Trivy scanning

`make scan`

