.PHONY: build generate test test-unit test-integration test-e2e lint coverage fmt clean migrate-up migrate-down migrate-status migrate-plan-down migrate-validate

VERSION ?= $(shell tr -d '\r\n' < ./cmd/server/VERSION)
LDFLAGS ?= -s -w -X main.Version=$(VERSION)
STEPS ?= 1

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -trimpath -o bin/server ./cmd/server

generate:
	go generate ./ent
	go generate ./cmd/server

test:
	go test ./...
	golangci-lint run ./...

lint:
	golangci-lint run ./...

coverage:
	go test -tags=unit -covermode=atomic -coverprofile=coverage-unit.out ./...
	go tool cover -func=coverage-unit.out | tail -n 1

fmt:
	go fmt ./...

clean:
	rm -rf bin coverage-unit.out

migrate-up:
	go run ./cmd/migrate -action up

migrate-down:
	go run ./cmd/migrate -action down -steps $(STEPS)

migrate-status:
	go run ./cmd/migrate -action status

migrate-plan-down:
	go run ./cmd/migrate -action plan-down -steps $(STEPS)

migrate-validate:
	go test ./migrations ./internal/repository ./cmd/migrate

test-unit:
	go test -tags=unit ./...

test-integration:
	go test -tags=integration ./...

test-e2e:
	./scripts/e2e-test.sh

test-e2e-local:
	go test -tags=e2e -v -timeout=300s ./internal/integration/...
