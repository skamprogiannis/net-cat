package server

import (
	"bufio"
	"net"
	"testing"
	"time"
)

// readResult keeps a line and its read error together so an async read can send
// one value through one channel instead of coordinating separate result streams.
type readResult struct {
	line string
	err  error
}

func setupTwoClients(t *testing.T) (*Client, *Client, net.Conn, net.Conn) {
	t.Helper()
	testClients, conns := setupBroadcastClients(t, "Patrick", "Bob")

	return testClients[0], testClients[1], conns[0], conns[1]
}

// setupBroadcastClients builds paired in-memory connections for broadcast
// tests. The returned Client values use the server side of each pipe, while
// the returned conns let the test read what each terminal-side client receives.
func setupBroadcastClients(t *testing.T, names ...string) ([]*Client, []net.Conn) {
	t.Helper()

	testClients := make([]*Client, 0, len(names))
	conns := make([]net.Conn, 0, len(names))

	deadline := time.Now().Add(time.Second)
	for _, name := range names {
		serverConn, clientConn := net.Pipe()
		serverConn.SetDeadline(deadline)
		clientConn.SetDeadline(deadline)

		testClients = append(testClients, &Client{
			conn:   serverConn,
			name:   name,
			writer: bufio.NewWriter(serverConn),
		})
		conns = append(conns, clientConn)

		t.Cleanup(func() {
			serverConn.Close()
			clientConn.Close()
		})
	}

	clients = testClients

	return testClients, conns
}

// readLineAsync starts the terminal-side read before broadcast writes. This
// keeps net.Pipe tests from blocking, because pipe writes wait for a reader.
func readLineAsync(conn net.Conn) <-chan readResult {
	ch := make(chan readResult, 1)
	go func() {
		line, err := bufio.NewReader(conn).ReadString('\n')
		ch <- readResult{line: line, err: err}
	}()
	return ch
}

// requireReadLine asserts that an async read produced want before the timeout.
func requireReadLine(t *testing.T, ch <-chan readResult, want string) {
	t.Helper()

	select {
	case result := <-ch:
		if result.err != nil {
			t.Fatalf("read broadcast: %v", result.err)
		}
		if result.line != want {
			t.Fatalf("broadcast = %q, want %q", result.line, want)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for broadcast")
	}
}

func TestBroadcast_SendsToAllReceivers(t *testing.T) {
	testClients, conns := setupBroadcastClients(t, "Patrick", "Bob", "George")

	bobRead := readLineAsync(conns[1])
	georgeRead := readLineAsync(conns[2])

	broadcastDone := make(chan struct{})
	go func() {
		broadcast("hello", testClients[0])
		close(broadcastDone)
	}()

	requireReadLine(t, bobRead, "hello\n")
	requireReadLine(t, georgeRead, "hello\n")

	<-broadcastDone
}

func TestBroadcast_SenderDoesNotGetMessage(t *testing.T) {
	testClients, conns := setupBroadcastClients(t, "Patrick", "Bob")

	bobRead := readLineAsync(conns[1])

	broadcastDone := make(chan struct{})
	go func() {
		broadcast("hello", testClients[0])
		close(broadcastDone)
	}()

	buf := make([]byte, 64)
	conns[0].SetDeadline(time.Now().Add(100 * time.Millisecond))
	n, _ := conns[0].Read(buf)
	if n > 0 {
		t.Fatalf("Patrick should not receive his own message")
	}

	requireReadLine(t, bobRead, "hello\n")
	<-broadcastDone
}
