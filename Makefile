APP_NAME := dms-api-server
CMD_PATH := ./cmd/server
BIN_PATH := ./bin/$(APP_NAME)
MIGRATIONS_DIR := ./migrations
DB_URL ?= postgres://postgres:postgres@localhost:5432/dms?sslmode=disable

.PHONY: help run build test tidy fmt clean migrate-up migrate-down migrate-down-all migrate-version migrate-create migrate-force

help:
	@echo "Available targets:"
	@echo "  make run    - Run the API server"
	@echo "  make build  - Build binary to $(BIN_PATH)"
	@echo "  make test   - Run tests"
	@echo "  make tidy   - Tidy module dependencies"
	@echo "  make fmt    - Format all Go files"
	@echo "  make clean  - Remove build artifacts"
	@echo "  make migrate-create NAME=<name>      - Create new SQL migration files"
	@echo "  make migrate-up                       - Apply pending migrations"
	@echo "  make migrate-down                     - Roll back one migration"
	@echo "  make migrate-down-all                 - Roll back all migrations"
	@echo "  make migrate-version                  - Show current migration version"
	@echo "  make migrate-force VERSION=<version>  - Force migration version"

run:
	go run $(CMD_PATH)

build:
	mkdir -p ./bin
	go build -o $(BIN_PATH) $(CMD_PATH)

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...

clean:
	rm -rf ./bin

migrate-create:
ifndef NAME
	$(error NAME is required. Example: make migrate-create NAME=create_orders_table)
endif
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

migrate-down-all:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down

migrate-version:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

migrate-force:
ifndef VERSION
	$(error VERSION is required. Example: make migrate-force VERSION=1)
endif
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(VERSION)
