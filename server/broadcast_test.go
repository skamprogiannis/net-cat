package server

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"
)

func setupTwoClients(t *testing.T) (*Client, *Client, net.Conn, net.Conn) {
	t.Helper()
	serverA, clientA := net.Pipe()
	serverB, clientB := net.Pipe()

	deadline := time.Now().Add(time.Second)
	serverA.SetDeadline(deadline)
	clientA.SetDeadline(deadline)
	serverB.SetDeadline(deadline)
	clientB.SetDeadline(deadline)

	patrick := &Client{conn: serverA, name: "Patrick", writer: bufio.NewWriter(serverA)}
	bob := &Client{conn: serverB, name: "Bob", writer: bufio.NewWriter(serverB)}
	clients = []*Client{patrick, bob}

	t.Cleanup(func() {
		serverA.Close()
		clientA.Close()
		serverB.Close()
		clientB.Close()
	})

	return patrick, bob, clientA, clientB
}

func TestBroadcast_ReceiverGetsMessage(t *testing.T) {
	patrick, _, _, clientB := setupTwoClients(t)

	done := make(chan struct{})
	go func() {
		broadcast("hello", patrick)
		close(done)
	}()

	buf := make([]byte, 64)
	n, err := clientB.Read(buf)
	if err != nil {
		t.Fatalf("Bob could not read: %v", err)
	}
	if !strings.Contains(string(buf[:n]), "hello") {
		t.Fatalf("expected 'hello', got '%s'", string(buf[:n]))
	}

	<-done
}

func TestBroadcast_SenderDoesNotGetMessage(t *testing.T) {
	patrick, _, clientA, clientB := setupTwoClients(t)

	readerDone := make(chan struct{})
	go func() {
		buf := make([]byte, 64)
		clientB.Read(buf)
		close(readerDone)
	}()

	broadcastDone := make(chan struct{})
	go func() {
		broadcast("hello", patrick)
		close(broadcastDone)
	}()

	buf := make([]byte, 64)
	clientA.SetDeadline(time.Now().Add(100 * time.Millisecond))
	n, _ := clientA.Read(buf)
	if n > 0 {
		t.Fatalf("Patrick should not receive his own message")
	}

	<-broadcastDone
	<-readerDone
}
