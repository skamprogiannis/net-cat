# Net-Cat

Net-Cat is a concurrent TCP chat server written in Go. It accepts standard
terminal clients such as Netcat, keeps a shared conversation history, and
coordinates up to ten connected users without data races.

Built collaboratively at Zone01 Athens, the project focuses on TCP networking,
goroutines, synchronized shared state, and integration testing with real
listeners.

## Highlights

- Serves multiple clients concurrently over TCP.
- Uses port `8989` by default or accepts one custom port argument.
- Prompts each client for a non-empty display name.
- Broadcasts timestamped messages, joins, and departures.
- Replays the in-memory event and message history to new clients.
- Rejects connections above the ten-client limit with a clear response.
- Handles disconnects without interrupting the remaining clients.
- Exercises concurrency under Go's race detector.

## Architecture

```text
main.go
  └─ resolve port and start server
       └─ server/server.go        accept loop and connection limit
            └─ server/handlers.go lifecycle for each client goroutine
                 ├─ server/prompts.go   welcome and name input
                 ├─ server/messages.go  formatting, history, and prompts
                 ├─ server/broadcast.go fan-out to connected clients
                 └─ server/state.go     synchronized shared state
```

Each accepted connection runs in its own goroutine. A mutex protects the
connected-client list, message history, connection count, and buffered writes
so concurrent messages do not corrupt shared state or interleave output.

## Requirements

- Go 1.24.3 or newer
- `nc` or another TCP client for manual testing

## Run

Start the server on the default port, `8989`:

```bash
go run .
```

Or supply one custom port:

```bash
go run . 2525
```

The server confirms the selected port:

```text
Listening on the port :8989
```

Connect from a separate terminal:

```bash
nc 127.0.0.1 8989
```

More than one port argument is rejected with:

```text
[USAGE]: ./TCPChat $port
```

See the [multi-client walkthrough](docs/demo.md) for a reproducible terminal
demo.

## Test

Run the unit and TCP integration tests, repeat them under the race detector,
and check the source with `go vet`:

```bash
go test ./...
go test -race ./...
go vet ./...
```

The suite covers port selection, the TCP welcome flow, the connection limit,
message formatting and history, broadcasts, join/leave behavior, blank input,
disconnect cleanup, and synchronized state access.

## Project Status

The Zone01 project requirements are implemented and covered by automated
tests. Net-Cat is intentionally a small educational server rather than a
production messaging platform:

- conversation history is held in memory and is lost on restart;
- clients share one chat room;
- TCP traffic is not encrypted or authenticated;
- display names do not have to be unique.

The original planning document is retained as
[historical project context](docs/project-plan.md). This README and the current
tests describe the implemented behavior.

## Team & Contributions

| Contributor | Focus |
| --- | --- |
| Stefanos Kamprogiannis | Welcome and history flows, disconnect handling, message delivery, TCP and concurrency tests, integration and documentation |
| Daniel Tymoshenko | Project foundation, TCP listener, initial client lifecycle, leave handling, and server logging |
| George Tzimokas | Connection limiting, name validation, broadcasting, shared-state safety, and related tests |
