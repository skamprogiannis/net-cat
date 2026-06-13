package server

import (
	"bufio"
	"io"
	"net"
	"testing"
	"time"
)

func TestReadClientName_ValidName(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	deadline := time.Now().Add(time.Second)
	serverConn.SetDeadline(deadline)
	clientConn.SetDeadline(deadline)

	go func() {
		clientConn.Write([]byte("George\n"))
	}()

	name, err := readClientName(serverConn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "George" {
		t.Fatalf("expected 'George', got '%s'", name)
	}
}
func TestReadClientName_EmptyThenValid(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	deadline := time.Now().Add(time.Second)
	serverConn.SetDeadline(deadline)
	clientConn.SetDeadline(deadline)

	go func() {
		buf := make([]byte, 64)
		clientConn.Write([]byte("\n"))
		clientConn.Read(buf) // διαβάζει το re-prompt
		clientConn.Write([]byte("   \n"))
		clientConn.Read(buf) // διαβάζει το re-prompt
		clientConn.Write([]byte("Gidian\n"))
	}()

	name, err := readClientName(serverConn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Gidian" {
		t.Fatalf("expected 'Gidian', got '%s'", name)
	}
}

func TestHandleClient_SkipsEmptyMessages(t *testing.T) {
	resetServerState(t)

	serverA, clientA := net.Pipe()
	serverB, clientB := net.Pipe()

	deadline := time.Now().Add(time.Second)
	serverA.SetDeadline(deadline)
	clientA.SetDeadline(deadline)
	serverB.SetDeadline(deadline)
	clientB.SetDeadline(deadline)

	bob := &Client{conn: serverB, name: "Bob", writer: bufio.NewWriter(serverB)}
	clients = []*Client{bob}
	connectionCount = 1

	t.Cleanup(func() {
		serverA.Close()
		clientA.Close()
		serverB.Close()
		clientB.Close()
	})

	done := make(chan struct{})
	go func() {
		handleClient(serverA)
		close(done)
	}()

	readFullWelcome(t, clientA)
	if _, err := clientA.Write([]byte("Patrick\n")); err != nil {
		t.Fatalf("write name: %v", err)
	}
	if _, err := clientA.Write([]byte("\n   \n  hello  \n")); err != nil {
		t.Fatalf("write messages: %v", err)
	}

	got := make([]byte, len("hello\n"))
	if _, err := io.ReadFull(clientB, got); err != nil {
		t.Fatalf("read broadcast: %v", err)
	}
	if string(got) != "hello\n" {
		t.Fatalf("expected only trimmed message, got %q", string(got))
	}

	clientA.Close()
	<-done
}

func resetServerState(t *testing.T) {
	t.Helper()

	clients = nil
	connectionCount = 0

	t.Cleanup(func() {
		clients = nil
		connectionCount = 0
	})
}

func readFullWelcome(t *testing.T, conn net.Conn) {
	t.Helper()

	want := welcomeMessage + namePrompt
	got := make([]byte, len(want))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatalf("read welcome: %v", err)
	}
	if string(got) != want {
		t.Fatalf("welcome = %q, want %q", string(got), want)
	}
}
