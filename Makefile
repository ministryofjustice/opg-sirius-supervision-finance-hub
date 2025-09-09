all:
	$(MAKE) test -j 3
	$(MAKE) build -j 3
	$(MAKE) scan -j 3
	$(MAKE) cypress
	$(MAKE) down


.PHONY: cypress zap

test: go-lint gosec unit-test test-migrations

test-results:
	mkdir -p -m 0777 test-results cypress/screenshots .trivy-cache .go-cache

setup-directories: test-results

go-lint:
	docker compose run --rm go-lint

gosec: setup-directories
	docker compose run --rm gosec

unit-test: setup-directories
	go run gotest.tools/gotestsum@latest --format testname  --junitfile test-results/unit-tests.xml -- ./... -coverprofile=test-results/test-coverage.txt

build: build-api build-hub build-migrations
build-api:
	docker compose build finance-api
build-hub:
	docker compose build finance-hub
build-migrations:
	docker compose build finance-migration

build-dev:
	docker compose -f docker-compose.yml -f docker/docker-compose.dev.yml build --parallel finance-hub finance-api yarn

build-all:
	docker compose build --parallel finance-hub finance-api finance-migration json-server cypress sirius-db allpay-mock holidays-api-mock

scan: scan-api scan-hub scan-migrations
scan-api: setup-directories
	docker compose run --rm trivy image --format table --exit-code 0 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-api:latest
	docker compose run --rm trivy image --format sarif --output /test-results/api.sarif --exit-code 1 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-api:latest
scan-hub: setup-directories
	docker compose run --rm trivy image --format table --exit-code 0 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-hub:latest
	docker compose run --rm trivy image --format sarif --output /test-results/hub.sarif --exit-code 1 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-hub:latest
scan-migrations: setup-directories
	docker compose run --rm trivy image --format table --exit-code 0 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-migration:latest
	docker compose run --rm trivy image --format sarif --output /test-results/migrations.sarif --exit-code 1 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-migration:latest
scan-allpay-mock: setup-directories
	docker compose run --rm trivy image --format table --exit-code 0 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-allpay-mock:latest
	docker compose run --rm trivy image --format sarif --output /test-results/migrations.sarif --exit-code 1 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-finance-allpay-mock:latest

clean:
	docker compose down
	docker compose run --rm yarn

up: clean build-dev start-and-seed sqlc-gen
	docker compose -f docker-compose.yml -f docker/docker-compose.dev.yml up finance-hub finance-api yarn

down:
	docker compose down

compile-assets:
	docker compose run --rm yarn build

sqlc-gen:
	docker compose run --rm sqlc generate

sqlc-diff:
	docker compose run --rm sqlc diff

sqlc-vet:
	docker compose run --rm sqlc vet

migrate:
	docker compose run --rm finance-migration

start-and-seed:
	docker compose up -d --wait sirius-db
	$(MAKE) build-migrations
	$(MAKE) migrate
	docker compose exec sirius-db psql -U user -d finance -a -f ./seed_data.sql
	docker compose up -d localstack json-server allpay-mock holidays-api-mock

test-migrations:
	docker compose pull finance-migration
	docker compose up -d --wait sirius-db
	$(MAKE) migrate
	docker compose exec sirius-db psql -U user -d finance -a -f ./seed_data.sql
	$(MAKE) build-migrations
	$(MAKE) migrate

cypress: setup-directories clean start-and-seed
	docker compose run cypress

export ACTIVE_SCAN ?= true
export ACTIVE_SCAN_TIMEOUT ?= 600
export SERVICE_NAME ?= FinanceHub
export SCAN_URL ?= http://finance-hub:8888/finance
cypress-zap: clean start-and-seed
	docker compose -f docker-compose.yml -f zap/docker-compose.zap.yml run --rm cypress
	docker compose -f docker-compose.yml -f zap/docker-compose.zap.yml exec -u root zap-proxy bash -c "apk add --no-cache jq"
	docker compose -f docker-compose.yml -f zap/docker-compose.zap.yml exec zap-proxy bash -c "/zap/wrk/scan.sh"
	docker compose -f docker-compose.yml -f zap/docker-compose.zap.yml down
