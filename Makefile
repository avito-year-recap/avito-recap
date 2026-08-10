
.PHONY: test test-render fmt docker-build up down

test:
	go test ./...

test-integration:
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

