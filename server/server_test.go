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

	if !strings.HasSuffix(welcomeMessage, "[ENTER YOUR NAME]: ") {
		t.Fatalf("welcomeMessage should end with name prompt on same line")
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

	got := make([]byte, len(welcomeMessage))
	if _, err := io.ReadFull(clientConn, got); err != nil {
		t.Fatalf("read welcome message: %v", err)
	}

	if string(got) != welcomeMessage {
		t.Fatalf("sendWelcomePrompt wrote %q, want %q", string(got), welcomeMessage)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("sendWelcomePrompt returned error: %v", err)
	}
}
