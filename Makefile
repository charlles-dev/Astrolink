.PHONY: install dev-node dev-portal dev-infra stop test test-node test-portal check build build-node build-portal lint clean

install:
	cd node && go mod download
	cd portal && npm install

dev-infra:
	docker compose -f docker-compose.dev.yml up -d

dev-node:
	cd node && go run ./cmd/server

dev-portal:
	cd portal && npm run dev -- --port 5173

stop:
	docker compose -f docker-compose.dev.yml down

test: test-node test-portal

test-node:
	cd node && go test ./...

test-portal:
	cd portal && npm test

check:
	cd portal && npm run check

build: build-node build-portal

build-node:
	New-Item -ItemType Directory -Force -Path node/dist
	cd node && go build -o dist/astrolink-node ./cmd/server

build-portal:
	cd portal && npm run build

lint:
	cd node && gofmt -w .
	cd node && go test ./...
	cd portal && npm run check

clean:
	Remove-Item -Recurse -Force node/dist -ErrorAction SilentlyContinue
	Remove-Item -Recurse -Force portal/.svelte-kit -ErrorAction SilentlyContinue
	Remove-Item -Recurse -Force portal/build -ErrorAction SilentlyContinue
