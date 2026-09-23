# ! /bin/bash

export PG_HOST="${PG_HOST:-localhost}"
export PG_PORT="${PG_PORT:-5432}"
export PG_SUPERUSER="${PG_SUPERUSER:-postgres}"
export PG_SUPERUSER_PASSWORD="${PG_SUPERUSER_PASSWORD:-password}"
export PG_SSL_MODE="${PG_SSL_MODE:-disable}"

export PG_DB_NAME="${PG_DB_NAME:-service_db}"
export PG_DB_SCHEMA="${PG_DB_SCHEMA:-service}"
export PG_DB_ROLE="${PG_DB_ROLE:-service}"
export PG_DB_ROLE_PASSWORD="${PG_DB_ROLE_PASSWORD:-password}"