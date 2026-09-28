platform:
	@. ./scripts/db/env.sh; \
	docker compose -f ./compose/docker-compose.local.yaml up -d

run:
	@. ./scripts/db/env.sh; \
	PG_CONNECTION_STRING=postgresql://$${PG_SUPERUSER}:$${PG_SUPERUSER_PASSWORD}@$${PG_HOST}:$${PG_PORT}/$${PG_DB_NAME}?sslmode=$${PG_SSL_MODE} \
	go run ./cmd/service/

generate:
	go generate ./...

# --- DATABASE ---
psql:
	@. ./scripts/db/env.sh; \
	psql postgresql://$${PG_SUPERUSER}:$${PG_SUPERUSER_PASSWORD}@$${PG_HOST}:$${PG_PORT}/$${PG_DB_NAME}?sslmode=$${PG_SSL_MODE}

init-db:
	@path=./scripts/db/init.sh; \
	chmod 0700 $$path; \
	"$$path"

name :=
migration:
	migrate create -ext .sql -dir ./migrations -seq -digits 4 ${name}

migrate-up:
	@. ./scripts/db/env.sh; \
	if ! command -v migrate >/dev/null 2>&1; then \
		echo "golang-migrate not installed. How to install: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate"; \
		exit 1; \
	fi; \
	migrate -path 'migrations' -database "postgresql://$${PG_SUPERUSER}:$${PG_SUPERUSER_PASSWORD}@$${PG_HOST}:$${PG_PORT}/$${PG_DB_NAME}?sslmode=$${PG_SSL_MODE}" up

migrate-down:
	@. ./scripts/db/env.sh; \
	if ! command -v migrate >/dev/null 2>&1; then \
		echo "golang-migrate not installed. How to install: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate"; \
		exit 1; \
	fi; \
	migrate -path 'migrations' -database "postgresql://$${PG_SUPERUSER}:$${PG_SUPERUSER_PASSWORD}@$${PG_HOST}:$${PG_PORT}/$${PG_DB_NAME}?sslmode=$${PG_SSL_MODE}" down