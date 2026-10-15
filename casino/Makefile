# Composes the platform that the casino needs
platform:
	@\
	networkID="$$(docker network ls --format='{{ .ID }}' --filter=name=casino-net)"; \
	if [[ -z "$$networkID" ]]; then \
		echo "Creating casino-net docker network..."; \
		docker network create casino-net; \
	fi
	@docker compose -f ./compose/docker-compose.platform.local.yaml up -d

# Decomposes the platform that the casino needs
platform-down:
	@docker compose -f ./compose/docker-compose.platform.local.yaml down

run-profile:
	@set -a; . .env; set +a; go run -tags=profile ./cmd/casino/

run:
	@set -a; . .env; set +a; go run ./cmd/casino/

generate:
	@go generate ./...

branch:=
update-pkg:
	go get github.com/Jordan-J-Brown/casino-go/pkg@${branch}

# --- Docker --- 

# Build the docker image for the casino. 
# Set NO_CACHE=1 to build without using cache.
# Set PROFILE=1 to build with profiling server enabled.
build: 
	@\
	flags=(\
		--tag casino:local \
		--build-context pkg=$$(realpath ../pkg) \
		--build-context gocache=$$(go env GOCACHE) \
		--build-context gomodcache=$$(go env GOMODCACHE) \
		--build-arg CONFIG_PATH=config.yaml \
		--file local.Dockerfile \
	); \
	if [[ -n "$$NO_CACHE" ]]; then \
		flags+=('--no-cache'); \
	fi; \
	if [[ -n "$$PROFILE" ]]; then \
		flags+=('--build-arg=BUILD_TAGS=profile'); \
	fi; \
	docker build "$${flags[@]}" .

up:
	@VERSION=local docker compose -f compose/docker-compose.casino.local.yaml up -d

# --- Swagger --- 

# Generates swagger docs from godoc comments using the swag-go tool
swag-init:
	@swag init \
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
	@set -a; . .env; set +a; \
	psql "$$PG_CONNECTION_STRING"

name :=
migration:
	@migrate create -ext .sql -dir ./migrations -seq -digits 4 ${name}

migrate-up:
	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "golang-migrate not installed. How to install: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate"; \
		exit 1; \
	fi; \
	set -a; . .env; set +a; \
	migrate -path 'migrations' -database "$$PG_CONNECTION_STRING" up

migrate-down:
	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "golang-migrate not installed. How to install: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate"; \
		exit 1; \
	fi; \
	set -a; . .env; set +a; \
	migrate -path 'migrations' -database "$$PG_CONNECTION_STRING" down -all