# TCPChat — NetCat Clone
## Product Requirements Document

**Team:** Daniel · George · Stefanos
**Project Duration:** April 21 – May 11, 2025
**Version:** 1.0

---

## Table of Contents

1. [What Are We Building?](#1-what-are-we-building)
2. [How It Works — The Big Picture](#2-how-it-works--the-big-picture)
3. [Key Concepts You Must Understand](#3-key-concepts-you-must-understand)
4. [Technical Stack & Rules](#4-technical-stack--rules)
5. [Team & Responsibilities](#5-team--responsibilities)
6. [Milestones & Detailed Task Breakdown](#6-milestones--detailed-task-breakdown)
7. [Functional Requirements](#7-functional-requirements)
8. [Acceptance Criteria](#8-acceptance-criteria)
9. [Out of Scope](#9-out-of-scope)
10. [Bonus Features](#10-bonus-features)

---

## 1. What Are We Building?

We are building **TCPChat** — a terminal-based group chat application written in Go. It is inspired by the Unix `nc` (NetCat) command-line tool, which is a low-level networking utility that reads and writes data across TCP or UDP network connections.

Our version is a fully functional **group chat server**. Think of it like a very basic version of IRC (Internet Relay Chat) — a server runs and listens for connections, and multiple users connect to it using their terminal and chat with each other in real time. We are only building the **server**. The client is simply the standard Unix `nc` tool that already exists on every Linux/Mac machine.

### What the end result looks like

The server is started like this:

```
$ go run .
Listening on the port :8989
```

A user connects using the standard `nc` tool from a different terminal:

```
$ nc localhost 8989
```

They are greeted with a Linux ASCII logo and asked for their name:

```
Welcome to TCP-Chat!
         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\dS"qML
 |    `.       | `' \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
     `-'       `--'
[ENTER YOUR NAME]:
```

Once they enter their name they are inside the chat. They can type messages, see messages from others, see who joins or leaves, and when they disconnect everyone else is notified. All messages are timestamped and labeled with the sender's name:

```
[2020-01-20 15:48:41][Daniel]: hey everyone
[2020-01-20 15:49:02][George]: what's up
George has left our chat...
```

---

## 2. How It Works — The Big Picture

Understanding the architecture before writing a single line of code is critical. Here is exactly how the system works from top to bottom.

### Server Side

The server is a Go program that:

1. Starts a **TCP listener** on a given port (default 8989)
2. Waits in a loop for incoming TCP connections
3. When a client connects, it spins up a **goroutine** (a lightweight concurrent thread) to handle that client independently
4. The goroutine sends the welcome message, collects the client's name, and then enters a read loop
5. Every message the client sends is **broadcast** to all other connected clients
6. When a client disconnects, the goroutine cleans up and notifies everyone else

### Client Side

The client is just a terminal running `nc` (NetCat). We do not write the client — it already exists as a Unix tool. Our server must be compatible with it.

### Data Flow of a Single Message

```
Client types "hello" and presses Enter
    → nc sends "hello\n" over TCP to the server
    → The client's goroutine on the server reads "hello"
    → Server checks: is it empty after trimming? No
    → Server formats: "[2020-01-20 15:48:41][Daniel]:hello"
    → Server appends formatted message to the history slice
    → Server loops over all connected clients
    → Server writes the formatted message to each client except the sender
    → All other clients see the message appear in their terminal instantly
```

### Shared State

The server maintains two critical pieces of shared state that all goroutines access simultaneously:

- **Client list** — a slice of all currently connected clients. Goroutines add to it when a client joins and remove from it when a client leaves.
- **Message history** — a slice of all formatted messages ever sent. New clients receive this on join so they can see the full conversation history.

Because multiple goroutines access these simultaneously, they must be protected with a **mutex** to prevent data races and crashes.

### What happens when a client joins

```
1. Server accepts TCP connection
2. Check: is connection count < 10? If no, reject and close
3. Send welcome message + Linux logo
4. Prompt for name, read name, validate it is not empty
5. Add client to global client list
6. Send full message history to new client
7. Broadcast "X has joined our chat..." to everyone else
8. Enter read loop — wait for messages from this client
```

### What happens when a client leaves

```
1. Read loop receives EOF or error — client disconnected
2. Remove client from global client list
3. Decrement connection counter
4. Close the TCP connection
5. Broadcast "X has left our chat..." to everyone remaining
6. Goroutine exits
```

---

## 3. Key Concepts You Must Understand

Every team member must be comfortable with these concepts before starting their tasks.

### TCP (Transmission Control Protocol)

TCP is a connection-oriented network protocol. When a client connects to our server, a persistent two-way connection is established. Both sides can send and receive data through this connection until one side closes it. In Go, a TCP connection is represented by the `net.Conn` interface. You can read from it and write to it like a file.

```go
// Starting a TCP listener
listener, err := net.Listen("tcp", ":8989")

// Accepting a connection
conn, err := listener.Accept()

// Writing to a connection
conn.Write([]byte("Hello!\n"))

// Reading from a connection
buf := make([]byte, 1024)
n, err := conn.Read(buf)
```

### Goroutines

A goroutine is a function that runs concurrently with other goroutines in the same program. In Go you launch one with the `go` keyword:

```go
go handleClient(conn)
```

This means `handleClient` runs in the background while the main loop immediately continues and waits for the next connection. Every client gets their own goroutine so one slow or disconnected client does not block the others.

### Mutexes (sync.Mutex)

A mutex (mutual exclusion lock) protects shared data from being read and written simultaneously by multiple goroutines, which would cause a **data race** — a bug where two goroutines access the same memory at the same time and produce unpredictable, incorrect results.

```go
var mu sync.Mutex
var clients []Client

// Safe write — lock before modifying, unlock after
mu.Lock()
clients = append(clients, newClient)
mu.Unlock()

// Safe read — also needs a lock
mu.Lock()
for _, c := range clients {
    // read client data
}
mu.Unlock()
```

Always unlock after locking. Use `defer mu.Unlock()` right after `mu.Lock()` to guarantee the unlock happens even if the function returns early or panics.

### bufio (Buffered I/O)

Reading from a TCP connection directly byte-by-byte is inefficient. `bufio.Reader` and `bufio.Writer` wrap the connection and buffer reads/writes internally, making them faster and easier to use. `bufio.Scanner` is particularly useful for reading line-by-line input:

```go
scanner := bufio.NewScanner(conn)
for scanner.Scan() {
    line := scanner.Text() // reads one line at a time
}
```

`bufio.Writer` is useful for writing — always call `.Flush()` after writing to ensure the data is actually sent over the network:

```go
writer := bufio.NewWriter(conn)
writer.WriteString("Hello!\n")
writer.Flush() // critical — without this the data stays in the buffer
```

### Error Handling in Go

Go does not use exceptions. Every function that can fail returns an `error` as its last return value. Always check errors:

```go
conn, err := listener.Accept()
if err != nil {
    log.Println("Accept error:", err)
    continue // don't crash, just skip this connection
}
```

---

## 4. Technical Stack & Rules

### Language

Go (Golang) — the entire project must be written in Go with no exceptions.

### Allowed packages only

No external libraries are permitted. Only these standard library packages:

| Package | Purpose |
|---------|---------|
| `net` | TCP connections and listeners |
| `sync` | Mutex for shared state protection |
| `time` | Timestamps for messages |
| `bufio` | Buffered reading/writing from connections |
| `fmt` | Formatting and printing |
| `log` | Server-side error and event logging |
| `os` | Reading command-line arguments |
| `strings` | String trimming and manipulation |
| `io` | Low-level I/O interfaces |
| `errors` | Error creation and handling |
| `reflect` | Reflection utilities if needed |

### Recommended project structure

```
TCPChat/
├── main.go          ← entry point, argument parsing, starts server
├── server.go        ← TCP listener, accept loop
├── client.go        ← Client struct, client list management
├── broadcast.go     ← broadcast function, message history
├── handlers.go      ← handleClient goroutine logic
└── *_test.go        ← unit test files
```

### Message format

Every chat message must follow this exact format:

```
[2006-01-02 15:04:05][name]:message
```

In Go code, generate the timestamp like this:

```go
timestamp := time.Now().Format("2006-01-02 15:04:05")
formatted := fmt.Sprintf("[%s][%s]:%s", timestamp, clientName, message)
```

### Port handling rules

| Scenario | Behavior |
|----------|---------|
| `go run .` | Use default port 8989 |
| `go run . 2525` | Use port 2525 |
| `go run . 2525 localhost` | Print usage message and exit |

Usage message must be exactly: `[USAGE]: ./TCPChat $port`

---

## 5. Team & Responsibilities

### Task Assignment Overview

| Task # | Task Name | Assigned To | Milestone |
|--------|-----------|-------------|-----------|
| #1 | Set up Go project structure & modules | **Daniel** | M1 |
| #2 | Implement TCP listener (default port 8989) | **Daniel** | M1 |
| #3 | Handle max 10 connections | **George** | M1 |
| #4 | Send Linux logo + name prompt on connect | **Stefanos** | M1 |
| #5 | Validate non-empty client name | **George** | M1 |
| #6 | Client struct + global client list | **Daniel** | M2 |
| #7 | Broadcast message to all clients | **George** | M2 |
| #8 | Block empty message broadcasting | **Stefanos** | M2 |
| #9 | Upload chat history to new client | **Stefanos** | M2 |
| #10 | Notify all on client join | **George** | M2 |
| #11 | Notify all on client leave | **Daniel** | M2 |
| #12 | Goroutine per client connection | **Daniel** | M3 |
| #13 | Mutex on shared state | **George** | M3 |
| #15 | Handle client disconnect gracefully | **Stefanos** | M3 |
| #16 | Server-side error logging | **Daniel** | M3 |
| #17 | Unit test: server connection | **Stefanos** | M4 |
| #18 | Unit test: message formatting | **George** | M4 |
| #19 | Unit test: broadcast logic | **Stefanos** | M4 |
| #20 | Edge case handling | **All** | M4 |

### Workload Summary

| Member | Tasks | Key Responsibilities |
|--------|-------|---------------------|
| **Daniel** | #1, #2, #6, #11, #12, #16, #20 | Foundation, TCP listener, client struct, goroutines, logging |
| **George** | #3, #5, #7, #10, #13, #18, #20 | Connections, name validation, broadcast, mutex safety, formatting test |
| **Stefanos** | #4, #8, #9, #15, #17, #19, #20 | Welcome prompt, history, disconnect handling, unit tests |

---

## 6. Milestones & Detailed Task Breakdown

---

### Milestone 1: Core Server Setup
**Deadline: April 25, 2025**

**Goal:** By the end of this milestone the server must be able to start, listen on a TCP port, handle command-line arguments correctly, accept incoming client connections, enforce the 10-connection limit, display the welcome message with the Linux logo, and collect a valid non-empty client name. No chat functionality yet — just a stable, connectable, named server.

---

#### Task #1 — Set up Go project structure & modules
**Assigned to: Daniel | Deadline: April 21**

This is the very first task and sets the foundation for the entire project. Daniel must initialize the Go module using `go mod init TCPChat` in the project root directory. This creates the `go.mod` file that defines the module name and Go version, which is required for the project to compile.

After initializing the module, Daniel must create the file structure. At minimum this means creating `main.go` as the entry point. It is strongly recommended to split the code into multiple files from the beginning — `server.go`, `client.go`, `broadcast.go`, `handlers.go` — so each team member can work on their own file without constant merge conflicts in Git.

Daniel must set up the shared Git repository and ensure all three team members can clone it and run `go run .` without errors. The initial `main.go` can simply print "Server starting..." and exit — the important thing is that the module compiles cleanly and the structure is agreed upon.

Establish basic coding conventions at this point: variable naming style (camelCase), how errors are handled and logged, where shared types like the `Client` struct will live, and how to use the `log` package consistently. Document these in a short `README.md` so everyone follows the same standards from day one.

---

#### Task #2 — Implement TCP listener (default port 8989)
**Assigned to: Daniel | Deadline: April 22**

Daniel must implement the core TCP listening logic. This involves two parts: argument parsing and starting the listener.

For argument parsing, use `os.Args` to read command-line arguments. `os.Args[0]` is always the program name itself. `os.Args[1]` is the port if provided. If `len(os.Args)` is greater than 2, the user passed too many arguments — print the usage message and exit immediately:

```go
if len(os.Args) > 2 {
    fmt.Println("[USAGE]: ./TCPChat $port")
    os.Exit(1)
}
port := "8989"
if len(os.Args) == 2 {
    port = os.Args[1]
}
```

Then start the TCP listener and print confirmation:

```go
listener, err := net.Listen("tcp", ":"+port)
if err != nil {
    log.Fatal(err)
}
defer listener.Close()
fmt.Println("Listening on the port :" + port)
```

After this, implement the accept loop — a `for` loop that calls `listener.Accept()` to block and wait for incoming connections. Each accepted `net.Conn` will eventually be passed to a goroutine (Task #12), but for now it can just be accepted, logged, and closed to verify the listener is working correctly. Never let an error from `Accept()` crash the server — log it and continue.

---

#### Task #3 — Handle max 10 connections
**Assigned to: George | Deadline: April 23**

The server must never allow more than 10 simultaneous client connections. George must implement a connection counter that is checked every time a new connection arrives in the accept loop.

Use a package-level integer variable to track current active connections. This variable is shared between all client goroutines, so it must always be read and written while holding a `sync.Mutex` lock. Every time a client is successfully accepted and registered, increment the counter. Every time a client disconnects (handled in Task #15), decrement it.

In the accept loop, immediately after receiving a new connection, lock the mutex, read the counter, and check if it is already at 10. If it is, write an informative message to the new connection, close it, unlock, and continue — do not increment the counter:

```go
mu.Lock()
if connectionCount >= 10 {
    mu.Unlock()
    conn.Write([]byte("Server is full. Maximum 10 connections reached.\n"))
    conn.Close()
    continue
}
connectionCount++
mu.Unlock()
```

It is critical that the decrement also happens correctly when clients leave. If the counter only goes up and never comes down, the server will eventually refuse all new connections even when slots are free. Coordinate with Stefanos (Task #15) to ensure the decrement is called on disconnect.

---

#### Task #4 — Send Linux logo + name prompt on connect
**Assigned to: Stefanos | Deadline: April 23**

When a new client connects and passes the connection limit check, the very first thing the server must do is send them the full welcome message. This is a fixed multi-line string that must be written to the client's connection immediately.

The welcome message starts with "Welcome to TCP-Chat!" followed by the Linux penguin ASCII art, and ends with "[ENTER YOUR NAME]: " on a new line. Store this as a constant string in the code and write it to the connection using `fmt.Fprintf(conn, welcomeMessage)` or a `bufio.Writer`.

The ASCII art must be preserved exactly as specified in the project requirements — whitespace and alignment matter because it is a visual diagram. Store it as a raw string literal using backticks in Go to avoid having to escape every special character. After writing the welcome message, flush the writer to ensure all bytes are sent over the network before waiting for the client's name input.

---

#### Task #5 — Validate non-empty client name
**Assigned to: George | Deadline: April 25**

After the welcome message and name prompt are sent, the server must read the client's response and validate it. George must implement a validation loop that keeps prompting the client until a valid non-empty name is received.

Use `bufio.NewReader(conn)` to read the client's input line by line. After reading a line, call `strings.TrimSpace()` on the result to remove any leading/trailing whitespace and newline characters. If the trimmed string is empty, send the prompt again and read again:

```go
reader := bufio.NewReader(conn)
var name string
for {
    name, err = reader.ReadString('\n')
    if err != nil {
        return // client disconnected during name entry
    }
    name = strings.TrimSpace(name)
    if name != "" {
        break
    }
    conn.Write([]byte("[ENTER YOUR NAME]: "))
}
```

Only after a valid name is confirmed should the client be registered in the global client list, the history sent to them, and the join notification broadcast. Handle the case where the client disconnects before entering a name — `ReadString` will return an error in that case and the goroutine should exit cleanly without registering the client.

---

### Milestone 2: Chat Functionality
**Deadline: May 1, 2025**

**Goal:** Implement everything related to the actual chat experience. Clients must be able to send messages that are formatted with timestamps and broadcast to all others. New clients must receive the full chat history on join. The server must announce when clients join or leave. Empty messages must be silently ignored. This is the most feature-rich milestone.

---

#### Task #6 — Client struct + global client list
**Assigned to: Daniel | Deadline: April 26**

Daniel must define the `Client` struct and the global data structures that the entire server depends on. This is one of the most important tasks because nearly every other task interacts with these types.

The `Client` struct must contain at minimum:
- `name string` — the client's display name
- `conn net.Conn` — the underlying TCP connection
- `writer *bufio.Writer` — a buffered writer for sending messages efficiently

```go
type Client struct {
    name   string
    conn   net.Conn
    writer *bufio.Writer
}
```

Define a package-level slice to hold all active clients, a package-level slice to hold message history, and a single `sync.Mutex` that protects both:

```go
var (
    clients []Client
    history []string
    mu      sync.Mutex
)
```

Implement two helper functions:
- `addClient(c Client)` — locks the mutex, appends to clients slice, unlocks
- `removeClient(conn net.Conn)` — locks the mutex, finds the client by connection, removes it from the slice, unlocks

```go
func removeClient(conn net.Conn) {
    mu.Lock()
    defer mu.Unlock()
    for i, c := range clients {
        if c.conn == conn {
            clients = append(clients[:i], clients[i+1:]...)
            return
        }
    }
}
```

---

#### Task #7 — Broadcast message to all clients
**Assigned to: George | Deadline: April 27**

George must implement the `broadcast` function — the core of the chat system. This function takes a formatted message string and a sender connection, then writes that message to every connected client except the sender.

The function must lock the mutex before iterating over the client list to prevent the list from being modified mid-iteration by another goroutine:

```go
func broadcast(message string, sender net.Conn) {
    mu.Lock()
    defer mu.Unlock()
    for _, c := range clients {
        if c.conn != sender {
            c.writer.WriteString(message + "\n")
            c.writer.Flush()
        }
    }
}
```

The message passed to this function must already be fully formatted with timestamp and name before `broadcast` is called. Handle write errors gracefully — if writing to a client fails, log the error but do not crash. A write failure usually means that client has already disconnected and their cleanup goroutine just hasn't run yet.

---

#### Task #8 — Block empty message broadcasting
**Assigned to: Stefanos | Deadline: April 27**

Inside the client's message read loop, after reading a line from the client, Stefanos must check whether the message is empty before doing anything with it. This check must happen before formatting, before adding to history, and before calling broadcast.

Use `strings.TrimSpace()` on the raw input to strip newlines and whitespace. If the result is an empty string, skip all processing and go back to waiting for the next message:

```go
for scanner.Scan() {
    message := strings.TrimSpace(scanner.Text())
    if message == "" {
        continue // silently ignore, do not broadcast
    }
    // proceed with formatting and broadcasting
}
```

This behavior must be completely silent — no error message is sent to the client, no log entry is created, nothing is broadcast. The client's terminal simply returns to the waiting state.

---

#### Task #9 — Upload chat history to new client
**Assigned to: Stefanos | Deadline: April 28**

When a new client joins the chat, they must immediately receive all messages that were sent before they connected. This allows them to read the conversation context before they start participating.

Maintain a global `history []string` slice (defined in Task #6) that stores every formatted message string in order. Every time a message is broadcast, append it to this slice first. When a new client has been validated and named, before broadcasting the join notification, send the entire history to the new client:

```go
mu.Lock()
for _, msg := range history {
    newClient.writer.WriteString(msg + "\n")
}
newClient.writer.Flush()
mu.Unlock()
```

This must happen while holding the mutex lock to prevent new messages from being appended to history mid-send. If the history is empty (first client ever), the loop simply does nothing and no errors occur.

---

#### Task #10 — Notify all on client join
**Assigned to: George | Deadline: April 29**

After a new client has been validated, added to the client list, and received the history, the server must notify all other currently connected clients that someone new has joined. The notification format is:

```
George has joined our chat...
```

This message is broadcast to all existing clients using the `broadcast` function. Importantly, this notification must **not** be added to the message history slice and must **not** include a timestamp — it is a system notification, not a chat message.

The order of operations when a new client joins must be strictly:
1. Validate and get the name
2. Add client to the client list
3. Send history to new client
4. Broadcast join notification to everyone else
5. Enter the message read loop

Getting this order wrong will cause messages to appear out of sequence for the new client.

---

#### Task #11 — Notify all on client leave
**Assigned to: Daniel | Deadline: May 1**

When a client disconnects for any reason — they pressed Ctrl+C, closed their terminal, or their network dropped — the server must notify all remaining clients that this person has left. The notification format is:

```
George has left our chat...
```

This is handled inside the client goroutine's cleanup logic. When the read loop exits, the cleanup sequence must run:

```go
name := client.name
removeClient(client.conn)

mu.Lock()
connectionCount--
mu.Unlock()

client.conn.Close()
broadcast(name + " has left our chat...", nil)
```

When broadcasting the leave message, the sender is `nil` because the leaving client's connection is already closed — passing `nil` means the message is sent to all remaining clients with no exclusion. Like the join notification, this message must not be added to history and must not include a timestamp.

---

### Milestone 3: Concurrency & Stability
**Deadline: May 6, 2025**

**Goal:** Harden the server against race conditions and unexpected failures. Every client runs in its own goroutine. All shared state is mutex-protected. Client disconnects are handled gracefully without affecting others. All events are properly logged. The server must pass `go run -race` with zero data race warnings.

---

#### Task #12 — Goroutine per client connection
**Assigned to: Daniel | Deadline: May 2**

Daniel must ensure every accepted client connection is handled in its own goroutine. In the accept loop, after a connection is accepted and the connection limit is checked, immediately launch a goroutine:

```go
for {
    conn, err := listener.Accept()
    if err != nil {
        log.Println("Accept error:", err)
        continue
    }
    go handleClient(conn)
}
```

The `handleClient` function manages the complete lifecycle of one client connection:

1. Send the welcome message and logo
2. Read and validate the client name
3. Create a `Client` struct and add it to the global list
4. Send message history to the new client
5. Broadcast the join notification
6. Enter a read loop that reads messages and broadcasts them
7. On loop exit, run cleanup: remove client, decrement counter, broadcast leave notification

Because each client runs in its own goroutine, the main accept loop is never blocked. A slow client, a client waiting for input, or a client that crashes cannot affect any other connected client.

---

#### Task #13 — Mutex on shared state
**Assigned to: George | Deadline: May 3**

George must audit the entire codebase and ensure every read and write to shared state is correctly protected by the `sync.Mutex`. The two shared resources are the `clients` slice and the `history` slice.

Go to every location in the code where either of these is accessed and verify it is wrapped in `mu.Lock()` / `mu.Unlock()`. Use `defer mu.Unlock()` immediately after `mu.Lock()` to guarantee the lock is always released:

```go
mu.Lock()
defer mu.Unlock()
// safe to read or write clients and history here
```

After completing this task, run the server with the race detector enabled:

```
go run -race .
```

Connect multiple clients simultaneously, send messages rapidly, and disconnect them abruptly. If the race detector prints any warnings, there is a missing mutex somewhere. Fix every single warning before marking this task complete. A server with data races will behave unpredictably and may crash under real load.

---

#### Task #15 — Handle client disconnect gracefully
**Assigned to: Stefanos | Deadline: May 4**

When a client disconnects, the server must handle it cleanly without panicking or affecting other clients. Stefanos must implement the disconnect detection and cleanup logic inside `handleClient`.

The read loop uses a `bufio.Scanner`. When the scanner returns `false` from `Scan()`, it means the connection was closed or an error occurred. After the loop exits, check `scanner.Err()` and log any non-nil error:

```go
for scanner.Scan() {
    // handle message
}
if err := scanner.Err(); err != nil {
    log.Println("Client read error:", err)
}
// cleanup runs here regardless of how the loop exited
```

Use `defer` at the start of `handleClient` to guarantee cleanup always runs no matter how the function exits:

```go
func handleClient(client Client) {
    defer func() {
        removeClient(client.conn)
        mu.Lock()
        connectionCount--
        mu.Unlock()
        client.conn.Close()
        broadcast(client.name + " has left our chat...", nil)
    }()
    // rest of handleClient...
}
```

This ensures that even if the function panics or returns early due to an error, the client is always removed from the list and remaining clients are always notified.

---

#### Task #16 — Server-side error logging
**Assigned to: Daniel | Deadline: May 6**

Daniel must ensure all significant server events and errors are logged consistently using the `log` package throughout the codebase. Using `fmt.Println` for errors is not acceptable — `log` automatically prepends a timestamp to every message and writes to stderr, making debugging far easier.

Log the following events at minimum:

```go
log.Println("Listening on port:", port)
log.Println("New connection from:", conn.RemoteAddr())
log.Printf("Client %s joined\n", name)
log.Printf("Client %s left\n", name)
log.Println("Accept error:", err)
log.Println("Write error for", name, ":", err)
log.Println("Read error for", name, ":", err)
```

Do not log every individual chat message — that would flood the output. Log structural events (connections, disconnections, errors) not message content. Never use `log.Fatal` inside a goroutine — it calls `os.Exit` and kills the entire server immediately. Only use `log.Fatal` in `main` for startup failures like a failed listener. Inside goroutines always use `log.Println` and handle the error gracefully without terminating.

---

### Milestone 4: Testing & Polish
**Deadline: May 9, 2025**

**Goal:** Write unit tests for core logic, manually verify all edge cases, clean up the code, and confirm the full end-to-end flow works correctly with real `nc` clients. The server must be stable, readable, and production-ready by the end of this milestone.

---

#### Task #17 — Unit test: server connection
**Assigned to: Stefanos | Deadline: May 7**

Stefanos must write a Go unit test in a `server_test.go` file that verifies the TCP listener starts correctly and accepts connections. Use the standard `testing` package.

The test should start the listener on a test port (e.g. 19999 to avoid conflicts with the default port), dial a connection to it, read the first response bytes, verify they contain the welcome message, then close everything:

```go
func TestServerAcceptsConnection(t *testing.T) {
    go startServer("19999")
    time.Sleep(100 * time.Millisecond)

    conn, err := net.Dial("tcp", "localhost:19999")
    if err != nil {
        t.Fatal("Could not connect:", err)
    }
    defer conn.Close()

    buf := make([]byte, 512)
    n, _ := conn.Read(buf)
    if !strings.Contains(string(buf[:n]), "Welcome to TCP-Chat") {
        t.Error("Expected welcome message, got:", string(buf[:n]))
    }
}
```

Run all tests with `go test ./...` from the project root. Make sure the server goroutine started in the test is properly shut down after the test completes to avoid port conflicts between test runs.

---

#### Task #18 — Unit test: message formatting
**Assigned to: George | Deadline: May 7**

George must write unit tests that verify the message formatting logic produces exactly the correct output. Extract the formatting into a standalone pure function so it can be tested in isolation without a live server:

```go
func formatMessage(name, message string, t time.Time) string {
    return fmt.Sprintf("[%s][%s]:%s", t.Format("2006-01-02 15:04:05"), name, message)
}
```

Write tests that pass in known inputs including a fixed timestamp and verify the output matches exactly:

```go
func TestMessageFormat(t *testing.T) {
    fixed := time.Date(2020, 1, 20, 15, 48, 41, 0, time.UTC)
    result := formatMessage("Daniel", "hello", fixed)
    expected := "[2020-01-20 15:48:41][Daniel]:hello"
    if result != expected {
        t.Errorf("Expected %q, got %q", expected, result)
    }
}
```

Also test edge cases: a name containing spaces, a message with special characters like `[]`, and a very long message. These help catch formatting bugs that only appear with unusual inputs.

---

#### Task #19 — Unit test: broadcast logic
**Assigned to: Stefanos | Deadline: May 8**

Stefanos must write a test that verifies the `broadcast` function correctly sends messages to all clients except the sender. Use `net.Pipe()` — a built-in Go function that creates an in-memory connected pair of `net.Conn` objects, perfect for testing without a real network:

```go
func TestBroadcast(t *testing.T) {
    conn1a, conn1b := net.Pipe()
    conn2a, conn2b := net.Pipe()

    clients = []Client{
        {name: "Alice", conn: conn1a, writer: bufio.NewWriter(conn1a)},
        {name: "Bob", conn: conn2a, writer: bufio.NewWriter(conn2a)},
    }

    broadcast("[timestamp][Alice]:hello", conn1a)

    buf := make([]byte, 64)
    conn2b.SetReadDeadline(time.Now().Add(time.Second))
    n, _ := conn2b.Read(buf)
    if !strings.Contains(string(buf[:n]), "hello") {
        t.Error("Bob did not receive the broadcast")
    }

    conn1b.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
    n, _ = conn1b.Read(buf)
    if n > 0 {
        t.Error("Alice should not receive her own message")
    }
}
```

Also write a test verifying that an empty message is never broadcast under any circumstances. Clean up all pipe connections with `defer conn.Close()`.

---

#### Task #20 — Edge case handling
**Assigned to: All (Daniel, George, Stefanos) | Deadline: May 9**

All three team members must collaboratively test and fix the following edge cases. Each person should start with the scenarios most related to their own tasks, then help the others.

**Edge cases to test and fix:**

1. **Empty name loop** — Connect with `nc`, press Enter repeatedly without typing a name. The server must re-prompt every time and never crash or register the client.

2. **11th connection attempt** — Connect 10 clients simultaneously and attempt an 11th. The 11th must receive a rejection message and be disconnected cleanly. The first 10 must be completely unaffected.

3. **Force disconnect mid-message** — While a client is connected and active, kill it with Ctrl+C. The server must detect the disconnect, clean up, and notify others without crashing or hanging.

4. **Simultaneous messages** — Have two or more clients send messages at the exact same time. Messages must not be garbled or mixed together. This tests whether the mutex is applied correctly everywhere.

5. **First client receives empty history** — The very first client to join should receive no history. Verify this produces no errors and no blank lines in the client's terminal.

6. **Reconnect after disconnect** — Disconnect a client and immediately reconnect with `nc`. The server must accept the new connection normally and the reconnected client must receive the full updated history.

Document every bug found with a short comment in the code explaining what the issue was and how it was fixed.

---

## 7. Functional Requirements

The following is the complete list of behaviors the server must implement. All items are required for the project to be considered complete.

- The server starts with `go run .` and listens on port 8989 by default
- If a port argument is provided, the server listens on that port instead
- If more than one argument is provided, the server prints the usage message and exits
- Upon connection, the client receives the welcome message with the Linux ASCII logo
- The client is prompted to enter a name before joining the chat
- Empty names are rejected and the client is re-prompted until a valid name is given
- The server enforces a maximum of 10 simultaneous connections
- When a client joins, all existing clients are notified
- When a client joins, they receive the full message history
- Messages are broadcast to all clients except the sender
- Messages include a timestamp and the sender's name in the correct format
- Empty messages are silently ignored and not broadcast
- When a client leaves, all remaining clients are notified
- Remaining clients stay connected when one client leaves
- The server does not crash under any normal usage scenario

---

## 8. Acceptance Criteria

The project is complete when all of the following pass:

- `go run .` starts the server on port 8989 with no errors
- `go run . 2525` starts the server on port 2525
- `go run . 2525 localhost` prints `[USAGE]: ./TCPChat $port` and exits immediately
- Two or more `nc` clients can connect, exchange messages, and disconnect without any server errors
- New clients see the full chat history immediately upon joining
- All messages appear in the format `[2006-01-02 15:04:05][name]:message`
- Empty messages produce no output on any client
- `go run -race .` produces zero data race warnings under concurrent load
- `go test ./...` passes all unit tests with no failures
- All 6 edge cases from Task #20 are handled without crashes

---

## 9. Out of Scope

The following are explicitly out of scope for the core project:

- UDP support
- Client-side application (the client is always `nc`)
- Authentication or password protection
- Message encryption
- Persistent database storage
- Web interface or REST API
- Any external Go packages not listed in the allowed packages section

---

## 10. Bonus Features

The following features are optional and must only be attempted after Milestones 1–4 are fully complete and stable. Do not work on these at the expense of the core implementation.

### Name change mid-session — Daniel
Allow clients to type `/name NewName` to change their display name while connected. The server must validate the new name is non-empty, update the `Client` struct in the global list under a mutex lock, and broadcast `"X has changed their name to Y"` to all other clients immediately.

### Log messages to file — Stefanos
On server start, open or create a `chat.log` file using `os.OpenFile` with append and create flags. Every broadcasted message must be written to this file in addition to being sent to clients. File writes must also be mutex-protected to avoid concurrent write corruption.

### Multiple chat rooms — George
Extend the server to support named chat rooms. Clients choose a room when they connect. Each room has its own client list and message history. Broadcasts are scoped to the room only. This requires restructuring the global state from flat slices into a `map[string]*Room` where each room holds its own clients and history slices.
