package pgdb

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	EnvPgConnectionString string = "PG_CONNECTION_STRING"
	EnvPgMaxConnections   string = "PG_MAX_CONNS"
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
	if maxConnsStr, ok := os.LookupEnv(EnvPgMaxConnections); ok {
		maxConns, err := strconv.Atoi(maxConnsStr)
		if err != nil {
			return nil, err
		}
		poolConfig.MaxConns = int32(maxConns)
	}

	return pgxpool.NewWithConfig(ctx, poolConfig)
}
