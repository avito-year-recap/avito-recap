.PHONY: generate proto-lint test run fmt

generate:
	npx --yes @bufbuild/buf@1.72.0 generate

proto-lint:
	npx --yes @bufbuild/buf@1.72.0 lint

test:
	go test ./...

run:
	go run ./cmd/api

fmt:
	gofmt -w ./cmd ./gen/go ./internal
