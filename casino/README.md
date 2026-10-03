# casino

## Running the casino locally

From within the casino directory:
1. Run `make platform` to compose core containers (NATS, Postgres etc).
2. Run `make migrate-up` to initialise the DB. 
3. Run `make run`.

## Database

- Create a migration with `make migration name="{migration_name}"`
- Use `make migrate-up` to run migrations
- Use `make migrate-down` to remove migrations
- Use `make psql` to connect to the service database from the terminal

# VSCode setup

Add these settings to your JSON so that integration test can run via VSCode:
```json
{
  "gopls": {
    "build.buildFlags": [
      "-tags=integration"
    ]
  },
  "go.testTags": "integration",
}
```