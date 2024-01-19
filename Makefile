all: go-lint unit-test build scan down

test-results:
	mkdir -p -m 0777 test-results cypress/screenshots .trivy-cache

setup-directories: test-results

go-lint:
	docker compose run --rm go-lint

build:
	docker compose build --parallel finance

unit-test: setup-directories
	docker compose run --rm test-runner gotestsum --junitfile test-results/unit-tests.xml -- ./... -coverprofile=test-results/test-coverage.txt

scan: setup-directories
	docker compose run --rm trivy image --format table --exit-code 0 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-supervision-finance-hub:latest
	docker compose run --rm trivy image --format sarif --output /test-results/trivy.sarif --exit-code 1 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/sirius-supervision-finance-hub:latest

up:
	docker compose up --build -d finance

dev-up:
	docker compose run --rm yarn
	docker compose -f docker-compose.yml -f docker/docker-compose.dev.yml build finance

down:
	docker compose down
