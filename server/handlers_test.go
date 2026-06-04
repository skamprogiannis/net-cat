package server

import (
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
