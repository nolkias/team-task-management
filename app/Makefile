.PHONY: build test test-unit test-integration cover migrate-up migrate-down

build:
	go build -o ./tmp/main ./cmd/api

test:
	go test ./...

test-unit:
	go test ./internal/...

test-integration:
	go test ./test/integration/...

cover:
	go test -coverprofile=cover.out ./internal/service/...
	go tool cover -func=cover.out

migrate-up:
	migrate -path migrations -database "mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)" up

migrate-down:
	migrate -path migrations -database "mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)" down
