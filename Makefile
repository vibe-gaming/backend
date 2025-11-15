include .env
export

LOCAL_BIN:=$(CURDIR)/bin
BUILD_DIR:=$(CURDIR)/cmd/app

# install goose migrator
install-goose:
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@latest

# lint
LINTER_VERSION=1.64.5
lint:
	@echo 'run golangci lint'
	@if ! $(GOPATH)/bin/golangci-lint --version | grep -q $(LINTER_VERSION); \
		then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v$(LINTER_VERSION); fi;
	$(GOPATH)/bin/golangci-lint run --out-format=tab -v --whole-files
	@echo

# compose deps
compose:
	@echo 'compose deps'
	docker compose -f docker-compose.yaml up -d

# down deps
compose-down:
	@echo 'compose deps'
	docker compose -f docker-compose.yaml down 

# build binary
build: deps build_binary

build-binary:
	@echo 'build backend binary'
	go build -o $(LOCAL_BIN) $(BUILD_DIR)

deps:
	@echo 'install dependencies'
	go mod tidy -v

# run app
run: deps run-app

run-app:
	@echo 'run backend'
	go run $(BUILD_DIR)/main.go

# generate swagger
swag:
	@echo 'generation swagger docs'
	swag init --parseDependency -g handler.go -dir internal/api/http/internal/v1 --instanceName internal

# migrations
LOCAL_MIGRATION_DIR=$(CURDIR)/migrations

LOCAL_MIGRATION_DSN="$(DB_USER):$(DB_PASSWORD)@tcp(localhost:3306)/$(DB_NAME)"

migration-status:
	$(LOCAL_BIN)/goose -dir ${LOCAL_MIGRATION_DIR} mysql ${LOCAL_MIGRATION_DSN} status -v

migration-up:
	$(LOCAL_BIN)/goose -dir ${LOCAL_MIGRATION_DIR} mysql ${LOCAL_MIGRATION_DSN} up -v

migration-down:
	$(LOCAL_BIN)/goose -dir ${LOCAL_MIGRATION_DIR} mysql ${LOCAL_MIGRATION_DSN} down -v

migration-create:
	 @echo "Migration name:"
	 @read migration_name; \
	 $(LOCAL_BIN)/goose -dir $(LOCAL_MIGRATION_DIR) create $$migration_name sql

migration-create:
	@echo "Migration name:"
	@read migration_name; \
	$(LOCAL_BIN)/goose -dir $(LOCAL_MIGRATION_DIR) create $$migration_name sql
