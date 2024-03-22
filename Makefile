all: go-lint unit-test build-all scan cypress down

.PHONY: cypress

test-results:
	mkdir -p -m 0777 test-results cypress/screenshots .trivy-cache

setup-directories: test-results

go-lint:
	docker compose run --rm go-lint

build:
	docker compose build --no-cache --parallel finance-hub finance-api

build-all:
	docker compose build --parallel finance-hub finance-api json-server test-runner cypress sirius-db

test: setup-directories
	go run gotest.tools/gotestsum@latest --format testname  --junitfile test-results/unit-tests.xml -- ./... -coverprofile=test-results/test-coverage.txt

scan: setup-directories
	docker compose run --rm trivy image --format table --exit-code 0 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-hub:latest
	docker compose run --rm trivy image --format sarif --output /test-results/trivy.sarif --exit-code 1 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-hub:latest

up:
	docker compose run --rm yarn
	docker compose -f docker-compose.yml -f docker/docker-compose.dev.yml build finance-hub finance-api
	docker compose -f docker-compose.yml -f docker/docker-compose.dev.yml up finance-hub yarn json-server finance-api sqlc sirius-db migrate

down:
	docker compose down

cypress: setup-directories
	docker compose up -d --wait finance-hub
	docker compose run --build --rm cypress

axe: setup-directories
	docker compose up -d --wait finance-hub
	docker compose run --rm cypress run --env grepTags="@axe"