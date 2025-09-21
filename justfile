
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

test-e2e-build:
	#!/usr/bin/env bash
	set -euo pipefail
	IMAGE_TAG="$(git rev-parse --short HEAD 2>/dev/null)-wip"

	if [[ "${VS_REBUILD:-false}" == "true" ]]; then
		echo "VS_REBUILD=true, rebuilding image $IMAGE_TAG..."
		docker build -t "versource-e2e:$IMAGE_TAG" -f Dockerfile .
		echo "Built image: $IMAGE_TAG"
	elif ! docker image inspect "versource-e2e:$IMAGE_TAG" >/dev/null 2>&1; then
		echo "Image $IMAGE_TAG not found, building..."
		docker build -t "versource-e2e:$IMAGE_TAG" -f Dockerfile .
		echo "Built image: $IMAGE_TAG"
	else
		echo "Image $IMAGE_TAG already exists, skipping build"
	fi

test-e2e: test-e2e-build test-prune
	go test -v -tags "e2e all" -count=1 ./tests/...

test-e2e-type type: test-e2e-build test-prune
	go test -v -tags "e2e {{type}}" -count=1 ./tests/...

test-e2e-datasets: test-e2e-build test-prune
	go test -v -tags "e2e datasets" -count=1 ./tests/...

test-prune:
	@docker system prune -f --volumes

data:
  cd data && docker compose up -d

migrate:
  go run main.go migrate

migrate-datasets: migrate test-e2e-datasets

serve: install
  versource serve

ui: install
  VS_CONFIG=local versource ui

ui-debug: install
  VS_CONFIG=local dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient -- ui

sql:
  dolt --host localhost --no-tls --use-db versource sql
