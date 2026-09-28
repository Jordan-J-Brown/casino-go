# casino-go

# Casino module
`./casino` is the player facing authenticator and Remote Game Server (RGS).

The casino uses NATS Request/Reply to handle communications with the engines. Engines use NATS QueueSubscribe, which allows horizontal scaling of engines. NATS will randomly select an engine to process the message.

`./internal/adapters` contains the HTTP API and NATS messaging API transport layer adapters.

`./internal/services` contains the core services which perform the core responsibilities of the casino, such as game flow, payouts and engine registrations.

## Engine modules

`./engines` contains the current engines implemented.

# pkg module

Contains shared library code and packages that all services can use. 

