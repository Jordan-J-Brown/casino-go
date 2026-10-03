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

Add this config to .vscode/launch.json:
```json
{
  {
    "name": "Run Casino",
    "type": "go",
    "request": "launch",
    "mode": "auto",
    "cwd": "${workspaceRoot}/casino",
    "program": "${workspaceRoot}/casino/cmd/service/main.go",
    "buildFlags": [ //These build flags are optional. They enable the profiling endpoints to be registered.
      "-tags",
      "local"
    ],
    "envFile": "/Users/jordanbrownwork/personal/casino-go/casino/.env"
  }
}
```

Add these settings to your JSON. This configures gopls correctly for the build tags this project uses:
```json
{
  "gopls": {
    "build.buildFlags": [
      "-tags=local,integration"
    ]
  },
  "go.testTags": "integration",
}
```