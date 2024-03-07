all: go-lint unit-test build scan cypress down

.PHONY: cypress

test-results:
	mkdir -p -m 0777 test-results cypress/screenshots .trivy-cache

setup-directories: test-results

go-lint:
	docker compose run --rm go-lint

build:
	docker compose build --no-cache --parallel finance

build-all:
	docker compose build --parallel finance json-server test-runner cypress

unit-test: setup-directories
	docker compose run --rm test-runner gotestsum --junitfile test-results/unit-tests.xml -- ./... -coverprofile=test-results/test-coverage.txt

scan: setup-directories
	docker compose run --rm trivy image --format table --exit-code 0 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-hub:latest
	docker compose run --rm trivy image --format sarif --output /test-results/trivy.sarif --exit-code 1 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-hub:latest

up:
	docker compose up --build -d finance

dev-up:
	docker compose run --rm yarn
	docker compose -f docker-compose.yml -f docker-compose.dev.yml build finance backend
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up finance yarn json-server sirius-db backend

down:
	docker compose down

cypress: setup-directories
	docker compose up -d --wait finance
	docker compose run --build --rm cypress

axe: setup-directories
	docker compose up -d --wait finance
	docker compose run --rm cypress run --env grepTags="@axe"