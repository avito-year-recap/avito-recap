.PHONY: generate proto-lint test run fmt

generate:
	npx --yes @bufbuild/buf@1.72.0 generate

proto-lint:
	npx --yes @bufbuild/buf@1.72.0 lint

.PHONY: test fmt docker-build up down

test:
	go test ./...

run:
	go run ./cmd/api

fmt:
	gofmt -w ./cmd ./internal

docker-build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down
