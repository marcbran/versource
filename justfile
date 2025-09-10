
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

test-e2e:
	go test -v -tags "e2e all" -count=1 ./tests/...

test-e2e-type type:
	go test -v -tags "e2e {{type}}" -count=1 ./tests/...

test-e2e-datasets:
	go test -v -tags "e2e datasets" -count=1 ./tests/...

test-coverage:
	go test -v -cover ./...

data:
  cd data && docker compose up -d

migrate:
  go run main.go migrate

serve: install
  versource serve

ui: install
  VS_CONFIG=local versource ui

ui-debug: install
  VS_CONFIG=local dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient -- ui

sql:
  dolt --host localhost --no-tls --use-db versource sql
