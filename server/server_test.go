package server

import (
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestWelcomeMessageContent(t *testing.T) {
	if !strings.HasPrefix(welcomeMessage, "Welcome to TCP-Chat!\n") {
		t.Fatalf("welcomeMessage should start with welcome line")
	}

	for _, line := range []string{
		"         _nnnn_",
		"        dGGGGMMb",
		"     `-'       `--'",
	} {
		if !strings.Contains(welcomeMessage, line) {
			t.Fatalf("welcomeMessage should contain logo line %q", line)
		}
	}

	if !strings.HasSuffix(welcomeMessage, "     `-'       `--'\n") {
		t.Fatalf("welcomeMessage should end with logo line")
	}
}

func TestSendWelcomePromptWritesFullMessage(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	deadline := time.Now().Add(time.Second)
	if err := serverConn.SetDeadline(deadline); err != nil {
		t.Fatalf("set server connection deadline: %v", err)
	}
	if err := clientConn.SetDeadline(deadline); err != nil {
		t.Fatalf("set client connection deadline: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- sendWelcomePrompt(serverConn)
	}()

	want := welcomeMessage + namePrompt
	got := make([]byte, len(want))
	if _, err := io.ReadFull(clientConn, got); err != nil {
		t.Fatalf("read welcome message: %v", err)
	}

	if string(got) != want {
		t.Fatalf("sendWelcomePrompt wrote %q, want %q", string(got), want)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("sendWelcomePrompt returned error: %v", err)
	}
}

func TestServeAcceptsConnectionAndSendsWelcome(t *testing.T) {
	resetServerState(t)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on test port: %v", err)
	}
	t.Cleanup(func() {
		listener.Close()
	})

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- serve(listener)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("dial test server: %v", err)
	}
	t.Cleanup(func() {
		conn.Close()
	})

	deadline := time.Now().Add(time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		t.Fatalf("set client deadline: %v", err)
	}

	want := welcomeMessage + namePrompt
	got := make([]byte, len(want))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatalf("read welcome from server: %v", err)
	}
	if string(got) != want {
		t.Fatalf("welcome = %q, want %q", string(got), want)
	}

	if err := conn.Close(); err != nil {
		t.Fatalf("close client connection: %v", err)
	}
	if err := listener.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	select {
	case err := <-serveDone:
		if err != nil {
			t.Fatalf("serve returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("serve did not stop after listener close")
	}

	waitForConnectionCount(t, 0)
}

func TestServe_RejectsEleventhConnection(t *testing.T) {
	resetServerState(t)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on test port: %v", err)
	}
	t.Cleanup(func() {
		listener.Close()
	})

	go serve(listener)

	// Fill all 10 slots; keep the connections open so they stay counted.
	conns := make([]net.Conn, 0, 10)
	for i := 0; i < 10; i++ {
		conn, err := net.Dial("tcp", listener.Addr().String())
		if err != nil {
			t.Fatalf("dial connection %d: %v", i, err)
		}
		conns = append(conns, conn)
	}

	// Wait until the server has registered all 10 before testing the 11th.
	waitForConnectionCount(t, 10)

	// The 11th connection must be rejected with the "full" message, not served.
	eleventh, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("dial 11th connection: %v", err)
	}
	eleventh.SetDeadline(time.Now().Add(time.Second))

	got, err := io.ReadAll(eleventh)
	if err != nil {
		t.Fatalf("read from 11th connection: %v", err)
	}
	want := "Server is full. Maximum 10 connections reached.\n"
	if string(got) != want {
		t.Fatalf("11th connection got %q, want %q", string(got), want)
	}

	// Drain everything and wait for the server to settle before returning, so
	// no leftover goroutine races with resetServerState's cleanup.
	eleventh.Close()
	for _, conn := range conns {
		conn.Close()
	}
	waitForConnectionCount(t, 0)
}

func waitForConnectionCount(t *testing.T, want int) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		got := connectionCount
		mu.Unlock()

		if got == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	mu.Lock()
	got := connectionCount
	mu.Unlock()
	t.Fatalf("connectionCount = %d, want %d", got, want)
}
