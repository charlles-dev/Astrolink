.PHONY: install dev dev-infra stop migrate test build lint clean

install:
	cd node && go mod download

dev-infra:
	docker compose -f docker-compose.dev.yml up -d

dev:
	cd node && go run ./cmd/server

stop:
	docker compose -f docker-compose.dev.yml down

migrate:
	@echo "Migrations are mounted into Postgres on first dev database startup."
	@echo "Use golang-migrate or psql for incremental migration execution in the next phase."

test:
	cd node && go test ./...

build:
	New-Item -ItemType Directory -Force -Path node/dist
	cd node && go build -o dist/astrolink-node ./cmd/server

lint:
	cd node && gofmt -w .
	cd node && go test ./...

clean:
	Remove-Item -Recurse -Force node/dist -ErrorAction SilentlyContinue
