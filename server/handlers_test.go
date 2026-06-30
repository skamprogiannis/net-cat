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
	joinLine := readIncomingLiveLine(t, reader)
	if joinLine != "Patrick has joined our chat...\n" {
		t.Fatalf("expected join notification first, got %q", joinLine)
	}

	if _, err := patrickConn.Write([]byte("\n   \n  hello  \n")); err != nil {
		t.Fatalf("write messages: %v", err)
	}

	line := readIncomingLiveLine(t, reader)
	if !strings.HasSuffix(line, "][Patrick]: hello\n") {
		t.Fatalf("expected only formatted trimmed message, got %q", line)
	}

	patrickConn.Close()
	leaveLine := readIncomingLiveLine(t, reader)
	if leaveLine != "Patrick has left our chat...\n" {
		t.Fatalf("got %q, want leave notification", leaveLine)
	}
	<-done
}

func TestHandleClient_BroadcastsMessageToSender(t *testing.T) {
	patrickConn, bobConn, done := startJoiningClientWithoutPromptDrain(t)

	bobReader := bufio.NewReader(bobConn)
	joinLine := readIncomingLiveLine(t, bobReader)
	if joinLine != "Patrick has joined our chat...\n" {
		t.Fatalf("expected join notification first, got %q", joinLine)
	}

	prompt := make([]byte, len(formatPrompt()))
	if _, err := io.ReadFull(patrickConn, prompt); err != nil {
		t.Fatalf("read Patrick prompt: %v", err)
	}
	if string(prompt) != formatPrompt() {
		t.Fatalf("expected Patrick prompt, got %q", string(prompt))
	}

	wantSenderLen := len(formatMessage("Patrick", "hello everyone", time.Now()) + "\n" + formatPrompt())
	patrickRead := readStringAsync(patrickConn, wantSenderLen)
	bobRead := readStringAsync(bobConn, wantSenderLen+1)

	if _, err := patrickConn.Write([]byte("hello everyone\n")); err != nil {
		t.Fatalf("write message: %v", err)
	}

	patrickOutput := requireReadLineValue(t, patrickRead)
	bobOutput := requireReadLineValue(t, bobRead)
	if bobOutput != "\n"+patrickOutput {
		t.Fatalf("sender got %q, receiver got %q", patrickOutput, bobOutput)
	}
	if !strings.Contains(patrickOutput, "][Patrick]: hello everyone\n> ") {
		t.Fatalf("expected formatted Patrick message and prompt, got %q", patrickOutput)
	}

	patrickConn.Close()
	leaveLine := readIncomingLiveLine(t, bufio.NewReader(bobConn))
	if leaveLine != "Patrick has left our chat...\n" {
		t.Fatalf("got %q, want leave notification", leaveLine)
	}
	<-done
}

func TestHandleClient_DoesNotBroadcastOnlyEmptyMessages(t *testing.T) {
	patrickConn, bobConn, done := startJoiningClient(t)

	reader := bufio.NewReader(bobConn)
	joinLine := readIncomingLiveLine(t, reader)
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

	leaveLine := readIncomingLiveLine(t, reader)
	if leaveLine != "Patrick has left our chat...\n" {
		t.Fatalf("got %q, want leave notification", leaveLine)
	}

	<-done
}

func TestHandleClient_SendsEventHistoryBeforePrompt(t *testing.T) {
	resetServerState(t)
	messageHistory = []string{
		"Alice has joined our chat...",
		"[2020-01-20 15:48:41][Alice]: hello",
		"Alice has left our chat...",
	}
	connectionCount = 1

	serverConn, clientConn := net.Pipe()
	deadline := time.Now().Add(time.Second)
	serverConn.SetDeadline(deadline)
	clientConn.SetDeadline(deadline)

	t.Cleanup(func() {
		serverConn.Close()
		clientConn.Close()
	})

	done := make(chan struct{})
	go func() {
		handleClient(serverConn)
		close(done)
	}()

	readFullWelcome(t, clientConn)
	if _, err := clientConn.Write([]byte("George\n")); err != nil {
		t.Fatalf("write name: %v", err)
	}

	want := "Alice has joined our chat...\n" +
		"[2020-01-20 15:48:41][Alice]: hello\n" +
		"Alice has left our chat...\n" +
		formatPrompt()
	got := make([]byte, len(want))
	if _, err := io.ReadFull(clientConn, got); err != nil {
		t.Fatalf("read history and prompt: %v", err)
	}
	if string(got) != want {
		t.Fatalf("history and prompt = %q, want %q", string(got), want)
	}

	clientConn.Close()
	<-done
}

func TestNotifyJoin(t *testing.T) {
	patrick, _, _, bobConn := setupTwoClients(t)

	done := make(chan struct{})
	go func() {
		notifyJoin(patrick)
		close(done)
	}()

	line := readIncomingLiveLine(t, bufio.NewReader(bobConn))
	want := "Patrick has joined our chat...\n"
	if line != want {
		t.Fatalf("got %q, want %q", line, want)
	}
	if len(messageHistory) != 1 || messageHistory[0] != "Patrick has joined our chat..." {
		t.Fatalf("history = %q, want join notification", messageHistory)
	}

	<-done
}

func TestHandleClient_NotifiesOnJoin(t *testing.T) {
	patrickConn, bobConn, done := startJoiningClient(t)

	line := readIncomingLiveLine(t, bufio.NewReader(bobConn))
	if line != "Patrick has joined our chat...\n" {
		t.Fatalf("got %q, want join notification", line)
	}

	patrickConn.Close()
	leaveLine := readIncomingLiveLine(t, bufio.NewReader(bobConn))
	if leaveLine != "Patrick has left our chat...\n" {
		t.Fatalf("got %q, want leave notification", leaveLine)
	}
	<-done
}

func TestNotifyLeaveStoresHistory(t *testing.T) {
	patrick, _, _, bobConn := setupTwoClients(t)

	done := make(chan struct{})
	go func() {
		notifyLeave(patrick)
		close(done)
	}()

	line := readIncomingLiveLine(t, bufio.NewReader(bobConn))
	want := "Patrick has left our chat...\n"
	if line != want {
		t.Fatalf("got %q, want %q", line, want)
	}
	if len(messageHistory) != 1 || messageHistory[0] != "Patrick has left our chat..." {
		t.Fatalf("history = %q, want leave notification", messageHistory)
	}

	<-done
}

// startJoiningClient registers Bob as an already-connected client, then drives
// a new client ("Patrick") through handleClient up to the moment he has joined.
// It returns Patrick's terminal-side conn (to send messages), Bob's
// terminal-side conn (to read what Bob receives), and a channel that is closed
// when handleClient returns.
func startJoiningClient(t *testing.T) (patrickConn, bobConn net.Conn, done chan struct{}) {
	t.Helper()
	patrickConn, bobConn, done = startJoiningClientWithoutPromptDrain(t)

	// The server sends each client its own prompt after joining and after every
	// message. On a synchronous net.Pipe those writes
	// block until someone reads, so drain Patrick's side in the background; the
	// tests using this helper only assert on what Bob receives.
	go func() {
		buf := make([]byte, 256)
		for {
			if _, err := patrickConn.Read(buf); err != nil {
				return
			}
		}
	}()

	return patrickConn, bobConn, done
}

// startJoiningClientWithoutPromptDrain registers Bob as an already-connected
// client and drives Patrick through the join flow without reading Patrick's
// post-join prompt. Tests that need to assert on Patrick's terminal output can
// read that prompt themselves before sending messages.
func startJoiningClientWithoutPromptDrain(t *testing.T) (patrickConn, bobConn net.Conn, done chan struct{}) {
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

func readIncomingLiveLine(t *testing.T, reader *bufio.Reader) string {
	t.Helper()

	prefix, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read live line prefix: %v", err)
	}
	if prefix != "\n" {
		t.Fatalf("live line prefix = %q, want newline", prefix)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read live line: %v", err)
	}

	prompt := make([]byte, len(formatPrompt()))
	if _, err := io.ReadFull(reader, prompt); err != nil {
		t.Fatalf("read prompt redraw: %v", err)
	}
	if string(prompt) != formatPrompt() {
		t.Fatalf("prompt redraw = %q, want %q", string(prompt), formatPrompt())
	}

	return line
}
