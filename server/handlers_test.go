package server

import (
	"bufio"
	"io"
	"net"
	"strings"
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
	patrickConn, bobConn, done := startJoiningClient(t)

	reader := bufio.NewReader(bobConn)
	joinLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read join notification: %v", err)
	}
	if joinLine != "Patrick has joined our chat...\n" {
		t.Fatalf("expected join notification first, got %q", joinLine)
	}

	if _, err := patrickConn.Write([]byte("\n   \n  hello  \n")); err != nil {
		t.Fatalf("write messages: %v", err)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read broadcast: %v", err)
	}
	if !strings.HasSuffix(line, "][Patrick]:hello\n") {
		t.Fatalf("expected only formatted trimmed message, got %q", line)
	}

	patrickConn.Close()
	<-done
}

func TestHandleClient_DoesNotBroadcastOnlyEmptyMessages(t *testing.T) {
	patrickConn, bobConn, done := startJoiningClient(t)

	reader := bufio.NewReader(bobConn)
	joinLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read join notification: %v", err)
	}
	if joinLine != "Patrick has joined our chat...\n" {
		t.Fatalf("expected join notification first, got %q", joinLine)
	}

	if _, err := patrickConn.Write([]byte("\n   \n\t\n")); err != nil {
		t.Fatalf("write empty messages: %v", err)
	}

	bobConn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	line, err := reader.ReadString('\n')
	if err == nil {
		t.Fatalf("expected no broadcast for empty messages, got %q", line)
	}
	if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
		t.Fatalf("expected timeout waiting for empty-message broadcast, got %v", err)
	}

	if err := bobConn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("reset Bob read deadline: %v", err)
	}
	patrickConn.Close()

	leaveLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read leave notification: %v", err)
	}
	if leaveLine != "Patrick has left our chat...\n" {
		t.Fatalf("got %q, want leave notification", leaveLine)
	}

	<-done
}

func TestNotifyJoin(t *testing.T) {
	patrick, _, _, bobConn := setupTwoClients(t)

	done := make(chan struct{})
	go func() {
		notifyJoin(patrick)
		close(done)
	}()

	line, err := bufio.NewReader(bobConn).ReadString('\n')
	if err != nil {
		t.Fatalf("read join notification: %v", err)
	}
	want := "Patrick has joined our chat...\n"
	if line != want {
		t.Fatalf("got %q, want %q", line, want)
	}

	<-done
}

func TestHandleClient_NotifiesOnJoin(t *testing.T) {
	patrickConn, bobConn, done := startJoiningClient(t)

	line, err := bufio.NewReader(bobConn).ReadString('\n')
	if err != nil {
		t.Fatalf("read join notification: %v", err)
	}
	if line != "Patrick has joined our chat...\n" {
		t.Fatalf("got %q, want join notification", line)
	}

	patrickConn.Close()
	<-done
}

// startJoiningClient registers Bob as an already-connected client, then drives
// a new client ("Patrick") through handleClient up to the moment he has joined.
// It returns Patrick's terminal-side conn (to send messages), Bob's
// terminal-side conn (to read what Bob receives), and a channel that is closed
// when handleClient returns.
func startJoiningClient(t *testing.T) (patrickConn, bobConn net.Conn, done chan struct{}) {
	t.Helper()
	resetServerState(t)

	patrickServer, patrickConn := net.Pipe()
	bobServer, bobConn := net.Pipe()

	deadline := time.Now().Add(time.Second)
	for _, c := range []net.Conn{patrickServer, patrickConn, bobServer, bobConn} {
		c.SetDeadline(deadline)
	}

	bob := &Client{conn: bobServer, name: "Bob", writer: bufio.NewWriter(bobServer)}
	clients = []*Client{bob}
	connectionCount = 1

	t.Cleanup(func() {
		patrickServer.Close()
		patrickConn.Close()
		bobServer.Close()
		bobConn.Close()
	})

	done = make(chan struct{})
	go func() {
		handleClient(patrickServer)
		close(done)
	}()

	readFullWelcome(t, patrickConn)
	if _, err := patrickConn.Write([]byte("Patrick\n")); err != nil {
		t.Fatalf("write name: %v", err)
	}

	return patrickConn, bobConn, done
}

func resetServerState(t *testing.T) {
	t.Helper()

	clients = nil
	messageHistory = nil
	connectionCount = 0

	t.Cleanup(func() {
		clients = nil
		messageHistory = nil
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
