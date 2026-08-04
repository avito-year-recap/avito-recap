.PHONY: test seed fmt

test:
	go test ./...

seed:
	go run ./cmd/seed

fmt:
	gofmt -w ./cmd ./internal
