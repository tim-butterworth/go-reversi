This is a multiplayer implementation of [reversi](https://en.wikipedia.org/wiki/Reversi) written in go.

This implementation is a command line version and all communication between the components is over direct tcp connections.

There are two components, a server and a client.

## To play

1) Start the server
    - from the project root directory run `go run cmd/server/main.go`
2) Start the client for player 1
    - open another terminal tab
    - from the project root directory run `go run cmd/client/main.go`
3) Start the client for player 2
    - open another terminal tab
    - from the project root directory run `go run cmd/client/main.go`

After the instructions have been completed, there should be 3 terminals open, one running the server, and two terminals each running the client.

The current implementation supports a single game, once the game is complete, everything has to be restarted.

ctrl + c will stop any of the processes, bringing down the server will cause the clients to terminate unless one of the clients is blocked waiting for user input, once user input in sprovided the client will terminate.

Features not yet supported but on the list:

- Ability to concede
- Reconnect to a game if connection is lost
- Start a new game once a game is complete with players on opposite sides