# Multi-client terminal demo

This walkthrough uses three terminals and the default port. It demonstrates a
join, a message delivered to both clients, history replay, and a clean leave.
Timestamps in your output will reflect the current time.

## 1. Start the server

In terminal 1, from the repository root:

```bash
go run .
```

Expected server output begins with:

```text
Listening on the port :8989
```

## 2. Connect Alice

In terminal 2:

```bash
nc 127.0.0.1 8989
```

After the welcome banner, enter `Alice` at the name prompt. The client displays
the chat prompt:

```text
[ENTER YOUR NAME]: Alice
>
```

Enter a message:

```text
hello from Alice
```

Alice receives the timestamped message back from the server:

```text
[YYYY-MM-DD HH:MM:SS][Alice]: hello from Alice
>
```

## 3. Connect Bob

In terminal 3:

```bash
nc 127.0.0.1 8989
```

Enter `Bob` at the name prompt. Bob receives the existing event and message
history before the chat prompt. Alice sees:

```text
Bob has joined our chat...
>
```

When Bob enters `hello Alice`, both clients receive:

```text
[YYYY-MM-DD HH:MM:SS][Bob]: hello Alice
>
```

## 4. Disconnect Bob

Press `Ctrl+C` in terminal 3. Alice receives:

```text
Bob has left our chat...
>
```

Press `Ctrl+C` in terminal 2 and then terminal 1 to stop the demo.
