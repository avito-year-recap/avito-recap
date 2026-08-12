
.PHONY: test test-integration test-all test-render fmt docker-build up down ollama-pull

test:
	go test ./...

test-integration:
	docker compose up -d clickhouse
	go test -tags=integration ./...

test-all:
	go test ./...
	go test -tags=integration ./...
	
test-render:
	go test ./internal/server -run RenderSingleService -count=1

fmt:
	gofmt -w ./cmd ./internal

docker-build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down


OLLAMA_MODEL ?= qwen3:4b

ollama-pull:
	ollama pull $(OLLAMA_MODEL)
