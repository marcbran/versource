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

snapshot-ci:
  #!/usr/bin/env bash
  set -euo pipefail
  goreleaser release --config .goreleaser.ci.yaml --snapshot --clean --skip archive

release:
  #!/usr/bin/env bash
  set -euo pipefail
  goreleaser release --config .goreleaser.yaml --clean --verbose

test-e2e: test-prune
  go test -v -tags "e2e all" -count=1 ./tests/...

test-e2e-ci:
  #!/usr/bin/env bash
  set -euo pipefail
  export IMAGE_TAG="SNAPSHOT-amd64"
  export VS_LOG=DEBUG
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

check-git-state:
  #!/usr/bin/env bash
  set -euo pipefail
  
  if [[ "$(git rev-parse --abbrev-ref HEAD)" != "main" ]]; then
    echo "Error: Must be on main branch"
    exit 1
  fi
  if [[ -n "$(git status --porcelain)" ]]; then
    echo "Error: Working directory is not clean"
    exit 1
  fi
  git fetch origin
  if [[ "$(git rev-parse HEAD)" != "$(git rev-parse origin/main)" ]]; then
    echo "Error: Local main is not in sync with origin/main"
    exit 1
  fi

bump-major: check-git-state
  #!/usr/bin/env bash
  set -euo pipefail
  
  LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
  VERSION=${LATEST_TAG#v}
  IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"
  
  NEW_VERSION="v$((MAJOR + 1)).0.0"
  
  git tag -a "$NEW_VERSION" -m "Bump version to $NEW_VERSION"
  git push origin "$NEW_VERSION"

bump-minor: check-git-state
  #!/usr/bin/env bash
  set -euo pipefail
  
  LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
  VERSION=${LATEST_TAG#v}
  IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"
  
  NEW_VERSION="v${MAJOR}.$((MINOR + 1)).0"
  
  git tag -a "$NEW_VERSION" -m "Bump version to $NEW_VERSION"
  git push origin "$NEW_VERSION"

bump-patch: check-git-state
  #!/usr/bin/env bash
  set -euo pipefail
  
  LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
  VERSION=${LATEST_TAG#v}
  IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"
  
  NEW_VERSION="v${MAJOR}.${MINOR}.$((PATCH + 1))"
  
  git tag -a "$NEW_VERSION" -m "Bump version to $NEW_VERSION"
  git push origin "$NEW_VERSION"