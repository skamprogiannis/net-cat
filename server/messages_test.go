package server

import (
	"bufio"
	"io"
	"net"
	"testing"
	"time"
)

func TestFormatMessage(t *testing.T) {
	fixedTime := time.Date(2020, 1, 20, 15, 48, 41, 0, time.UTC)

	got := formatMessage("Daniel", "hello", fixedTime)
	want := "[2020-01-20 15:48:41][Daniel]:hello"

	if got != want {
		t.Fatalf("formatMessage() = %q, want %q", got, want)
	}
}

func TestAddMessageToHistory(t *testing.T) {
	resetServerState(t)

	addMessageToHistory("[2020-01-20 15:48:41][Daniel]:hello")

	if len(messageHistory) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(messageHistory))
	}
	if messageHistory[0] != "[2020-01-20 15:48:41][Daniel]:hello" {
		t.Fatalf("unexpected history entry %q", messageHistory[0])
	}
}

func TestSendHistory(t *testing.T) {
	resetServerState(t)

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	deadline := time.Now().Add(time.Second)
	serverConn.SetDeadline(deadline)
	clientConn.SetDeadline(deadline)

	messageHistory = []string{
		"[2020-01-20 15:48:41][Daniel]:hello",
		"[2020-01-20 15:49:02][George]:what's up",
	}
	client := &Client{
		conn:   serverConn,
		name:   "Stefanos",
		writer: bufio.NewWriter(serverConn),
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- sendHistory(client)
	}()

	want := "[2020-01-20 15:48:41][Daniel]:hello\n" +
		"[2020-01-20 15:49:02][George]:what's up\n"
	got := make([]byte, len(want))
	if _, err := io.ReadFull(clientConn, got); err != nil {
		t.Fatalf("read history: %v", err)
	}
	if string(got) != want {
		t.Fatalf("history = %q, want %q", string(got), want)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("sendHistory returned error: %v", err)
	}
}
