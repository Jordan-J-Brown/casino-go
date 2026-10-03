pgConnectionString := postgres://postgres:password@localhost:5432/postgres?sslmode=disable

# Composes the platform that the casino needs
platform:
	@. ./scripts/db/env.sh; \
	docker compose -f ./compose/docker-compose.platform.local.yaml up -d

# Decomposes the platform that the casino needs
platform-down:
	docker compose -f ./compose/docker-compose.platform.local.yaml down

# Runs the casino service locally via go run
run:
	@. ./scripts/db/env.sh; \
	PG_CONNECTION_STRING="${pgConnectionString}" \
	go run ./cmd/service/

generate:
	go generate ./...

# Build the docker image for the casino service
tag :=
build: 
	@GITHUB_TOKEN=$$GITHUB_TOKEN \
	docker build --tag casino:${tag} \
		--secret id=GITHUB_TOKEN \
		--build-arg CONFIG_PATH=configs/config.prod.yaml .

# --- Testing ---
test:
	@go test -tags !integrations ./...

test-integrations:
	@go test -tags integrations ./...

# --- DATABASE ---

psql:
	@. ./scripts/db/env.sh; \
	psql "${pgConnectionString}"

name :=
migration:
	migrate create -ext .sql -dir ./migrations -seq -digits 4 ${name}

migrate-up:
	@. ./scripts/db/env.sh; \
	if ! command -v migrate >/dev/null 2>&1; then \
		echo "golang-migrate not installed. How to install: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate"; \
		exit 1; \
	fi; \
	migrate -path 'migrations' -database "${pgConnectionString}" up

migrate-down:
	@. ./scripts/db/env.sh; \
	if ! command -v migrate >/dev/null 2>&1; then \
		echo "golang-migrate not installed. How to install: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate"; \
		exit 1; \
	fi; \
	migrate -path 'migrations' -database "${pgConnectionString}" down -all