# ! /bin/bash
source './scripts/db/env.sh'

if [[ -z "$(command -v migrate)" ]]; then
  echo "golang-migrate not installed. How to install: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate"
  exit 1
fi

if [[ -z "$(command -v psql)" ]]; then 
  echo "psql is not installed."
  exit 1
fi

if [[ -z "$PG_DB_ROLE_PASSWORD" ]]; then
  echo "PG_DB_ROLE_PASSWORD is not set"
  exit 1
fi

PSQL=(psql postgresql://${PG_SUPERUSER}:${PG_SUPERUSER_PASSWORD}@${PG_HOST}:${PG_PORT}/postgres) 
PSQL_DB=(psql postgresql://${PG_SUPERUSER}:${PG_SUPERUSER_PASSWORD}@${PG_HOST}:${PG_PORT}/${PG_DB_NAME}) 

# Check if database exists
DB_EXISTS=$("${PSQL[@]}" -tAc "SELECT 1 FROM pg_database WHERE datname = '${PG_DB_NAME}';")
echo "DB_EXISTS: $DB_EXISTS"
if [[ -z "$DB_EXISTS" ]]; then
  echo "$PG_DB_NAME database does not exist, creating..."
  "${PSQL[@]}" -c "CREATE DATABASE $PG_DB_NAME;"
else 
  echo "$PG_DB_NAME database exists, skipping creation..."
fi

ROLE_EXISTS=$("${PSQL[@]}" -tAc "SELECT 1 FROM pg_roles WHERE rolname='${PG_DB_ROLE};'")
echo "ROLE_EXISTS: $ROLE_EXISTS"
if [[ -z "$DB_EXISTS" ]]; then
  echo "Role '$PG_DB_ROLE' does not exist, creating..."
  "${PSQL[@]}" -c "CREATE ROLE $PG_DB_ROLE WITH LOGIN PASSWORD '$PG_DB_ROLE_PASSWORD' NOSUPERUSER NOCREATEDB NOCREATEROLE;"
else 
  echo "Role '$PG_DB_ROLE' exists, skipping creation..."
fi

"${PSQL_DB[@]}" <<SQL
CREATE SCHEMA IF NOT EXISTS ${PG_DB_SCHEMA};

GRANT CONNECT ON DATABASE ${PG_DB_NAME} TO ${PG_DB_ROLE};
GRANT USAGE ON SCHEMA ${PG_DB_SCHEMA} TO ${PG_DB_ROLE};
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA ${PG_DB_SCHEMA} TO ${PG_DB_ROLE};
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA ${PG_DB_SCHEMA} TO ${PG_DB_ROLE};
ALTER DEFAULT PRIVILEGES IN SCHEMA ${PG_DB_SCHEMA} GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ${PG_DB_ROLE};
ALTER DEFAULT PRIVILEGES IN SCHEMA ${PG_DB_SCHEMA} GRANT USAGE, SELECT ON SEQUENCES TO ${PG_DB_ROLE};
SQL

echo "DB initialised.."

# Run migrations
echo "Running migrations..."
migrate -path 'migrations' -database "postgresql://${PG_SUPERUSER}:${PG_SUPERUSER_PASSWORD}@${PG_HOST}:${PG_PORT}/${PG_DB_NAME}?sslmode=${PG_SSL_MODE}" up