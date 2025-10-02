set export

default:
  @just --list

build:
  go build -v ./...

install: build
  go install -v ./...

lint:
  golangci-lint run

test:
  go test -v ./...

test-coverage:
  go test -v -cover ./...

snapshot:
  #!/usr/bin/env bash
  set -euo pipefail

  goreleaser release --config .goreleaser.dev.yaml --snapshot --clean --skip archive

release:
  #!/usr/bin/env bash
  set -euo pipefail

  goreleaser release --config .goreleaser.yaml --clean

test-e2e: test-prune
  go test -v -tags "e2e all" -count=1 ./tests/...

test-e2e-type type: test-prune
  go test -v -tags "e2e {{type}}" -count=1 ./tests/...

test-e2e-datasets: test-prune
  go test -v -tags "e2e datasets" -count=1 ./tests/...

test-prune:
  docker system prune -f --volumes

sql:
  dolt --host localhost --no-tls --use-db versource sql

migrate: install
  VS_CONFIG=local versource migrate

migrate-datasets: migrate test-e2e-datasets

serve: install
  VS_CONFIG=local versource serve

ui: install
  VS_CONFIG=local versource ui

ui-debug: install
  VS_CONFIG=local dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient -- ui


export:
  TEST="asdf"
  export TEST1="qwer"

echo: export
  echo ${TEST:-""}
  echo ${TEST1:-""}
