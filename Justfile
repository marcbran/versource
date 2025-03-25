
default:
    @just --list

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run
