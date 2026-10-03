# Casino-Go

## Project overview

This project is an online casino/Remote Game Server (RGS) built in Golang. Engines connect to the casino service via NATS.

### Disclaimer

This is just a portfolio/hobby project and does not represent a production worthy RGS. It serves as a place to demonstrate the knowledge and skills for building backend services using Golang.

### Features
- REST API to authenticate and play games.
- Swagger docs and swagger endpoint.
- Postgres data store for players and games.
- Microservice based engines that communicate with the casino via NATS queue groups. Queue groups allow horizontal scaling of engines, picking one randomly to process the messages.
- Integration tests utilising testcontainers, isolated using go build tags.

### Future improvements 
- Multiplayer games using web sockets (poker, blackjack).
- Move bet size and max payout multiplier configs to Postgres, along with an in-memory cache for fast lookups.
- Implement real authentication handling. Currently the authentication endpoint just creates a new player with a balance to play the engines.

## Running the casino locally

From within the casino directory:
1. Run `make platform` to compose core containers (NATS, Postgres etc).
2. Run `make migrate-up` to initialise the DB. 
3. Run `make run`.
4. Within the `./engines/mines` directory, run `make run` to start the mines engine.

## Database

- Create a migration with `make migration name="{migration_name}"`
- Use `make migrate-up` to run migrations
- Use `make migrate-down` to remove migrations
- Use `make psql` to connect to the service database from the terminal

## VSCode setup

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

## Swagger 

This project uses swag-go to generate swagger docs. Models that contain a field of "string" type but have a format of "json" are actually JSON objects. This is a limitation of the swag-go library it seems. A Postman collection of API routes can be found in [API.postman_collection.json](./API.postman_collection.json).