pgConnectionString := postgres://postgres:password@localhost:5432/postgres?sslmode=disable

# Composes the platform that the casino needs
platform:
	@docker compose -f ./compose/docker-compose.platform.local.yaml up -d

# Decomposes the platform that the casino needs
platform-down:
	docker compose -f ./compose/docker-compose.platform.local.yaml down

run-profile:
	@set -a; . .env; set +a; go run -tags=profile ./cmd/casino/

run:
	@set -a; . .env; set +a; go run ./cmd/casino/

generate:
	go generate ./...

build: 
	@GITHUB_TOKEN=$$GITHUB_TOKEN \
	docker build --tag casino:local \
		--secret id=GITHUB_TOKEN \
		--build-arg CONFIG_PATH=configs/config.prod.yaml .

compose-casino:
	docker compose -f compose/docker-compose.casino.local.yaml up -d

# Generates swagger docs from godoc comments using the swag-go tool
swag-init:
	swag init \
		--parseInternal \
		--parseDependency \
		--dir ./internal/adapters/web,pkg

# --- Testing ---
test:
	@go test -tags !integration ./...

test-integration:
	@go test -tags integration ./...

# --- DATABASE ---

psql:
	@psql "${pgConnectionString}"

name :=
migration:
	migrate create -ext .sql -dir ./migrations -seq -digits 4 ${name}

migrate-up:
	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "golang-migrate not installed. How to install: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate"; \
		exit 1; \
	fi; \
	migrate -path 'migrations' -database "${pgConnectionString}" up

migrate-down:
	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "golang-migrate not installed. How to install: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate"; \
		exit 1; \
	fi; \
	migrate -path 'migrations' -database "${pgConnectionString}" down -all