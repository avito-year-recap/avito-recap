
.PHONY: test fmt docker-build up down

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

docker-build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

