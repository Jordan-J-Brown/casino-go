# casino

## Running the casino locally

From within the casino directory:
1. Run `make platform` to compose core containers (NATS, Postgres etc).
2. Run `make init-db` to set up service DB user and run migrations on the DB.
3. Run `make run`.

## Database

- Create a migration with `make migration name="{migration_name}"`
- Use `make migrate-up` to run migrations
- Use `make migrate-down` to remove migrations
- Use `make psql` to connect to the service database from the terminal