all: go-lint test build-all scan cypress axe down

.PHONY: cypress

test-results:
	mkdir -p -m 0777 test-results cypress/screenshots .trivy-cache .go-cache

setup-directories: test-results

go-lint:
	docker compose run --rm go-lint

build:
	docker compose build --no-cache --parallel finance-hub finance-api finance-migration

build-dev:
	docker compose -f docker-compose.yml -f docker/docker-compose.dev.yml build --parallel finance-hub finance-api yarn

build-all:
	docker compose build --parallel finance-hub finance-api finance-migration json-server cypress sirius-db

test: setup-directories
	go run gotest.tools/gotestsum@latest --format testname  --junitfile test-results/unit-tests.xml -- ./... -coverprofile=test-results/test-coverage.txt

scan: setup-directories
	docker compose run --rm trivy image --format table --exit-code 0 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-hub:latest
	docker compose run --rm trivy image --format sarif --output /test-results/hub.sarif --exit-code 1 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-hub:latest
	docker compose run --rm trivy image --format table --exit-code 0 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-api:latest
	docker compose run --rm trivy image --format sarif --output /test-results/api.sarif --exit-code 1 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-api:latest
	docker compose run --rm trivy image --format table --exit-code 0 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-migration:latest
	docker compose run --rm trivy image --format sarif --output /test-results/migration.sarif --exit-code 1 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-migration:latest

clean:
	docker compose down
	docker compose run --rm yarn

up: clean build-dev start-and-seed sqlc-gen
	docker compose -f docker-compose.yml -f docker/docker-compose.dev.yml up finance-hub json-server finance-api yarn localstack

down:
	docker compose down

compile-assets:
	docker compose run --rm yarn build

sqlc-gen:
	docker compose run --rm sqlc-gen

sqlc-diff:
	docker compose run --rm sqlc-diff

sqlc-vet:
	docker compose run --rm sqlc-vet

start-and-seed:
	docker compose up -d --wait sirius-db
	docker compose run --rm --build finance-migration
	docker compose exec sirius-db psql -U user -d finance -a -f ./seed_data.sql

cypress: setup-directories clean start-and-seed
	docker compose run --build cypress
