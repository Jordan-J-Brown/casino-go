package pgdb

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	EnvPgConnectionString string = "PG_CONNECTION_STRING"
)

// Retrieves the connection string from PG_CONNECTION_STRING and initialises
// a postgres connection pool.
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	connectionString, ok := os.LookupEnv(EnvPgConnectionString)
	if !ok {
		return nil, errors.New(EnvPgConnectionString + " is missing")
	}

	poolConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, err
	}

	//TODO: setup sensible defaults

	return pgxpool.NewWithConfig(ctx, poolConfig)
}
