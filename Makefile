MIGRATE := migrate
DB_URL  ?= postgres://appraisal:appraisal@localhost:5433/notification_db?sslmode=disable

.PHONY: build run test migrate-up migrate-down

build:
	go build ./...

run:
	go run cmd/server/main.go

test:
	go test ./...

migrate-up:
	$(MIGRATE) -path migrations -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path migrations -database "$(DB_URL)" down 1
