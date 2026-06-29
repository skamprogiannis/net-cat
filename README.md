# TCPChat

TCPChat is a small Go TCP chat server inspired by NetCat. Users connect with
`nc`, enter a name, and chat with everyone connected to the same server.

## Requirements

- Go 1.24.3 or newer
- `nc` or another TCP client for manual testing

## Run

Start the server with a port:

```bash
go run . 8989
```

Connect from another terminal:

```bash
nc localhost 8989
```

The server prints:

```text
Listening on the port :8989
```

Running without exactly one port argument prints:

```text
[USAGE]: ./TCPChat $port
```

## Behavior

- Sends a welcome banner and asks each client for a non-empty name.
- Keeps up to 10 active connections.
- Broadcasts join and leave notifications to other clients.
- Formats chat messages as `[YYYY-MM-DD HH:MM:SS][name]:message`.
- Sends existing message history to clients when they join.
- Skips empty or whitespace-only client messages.
- Removes disconnected clients without stopping other active clients.

## Test

The test suite includes unit tests and TCP listener coverage.

```bash
go test ./...
go test -race ./...
```
