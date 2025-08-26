
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
	go test -v -tags e2e -count=1 ./tests/...

test-coverage:
	go test -v -cover ./...

data:
  cd data && docker compose up -d

migrate:
  go run main.go migrate

serve: install
  versource serve

sql:
  dolt --host localhost --no-tls --use-db versource sql
