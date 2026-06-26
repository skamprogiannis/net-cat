package server

import (
	"sync"
	"testing"
)

func TestConcurrentSharedStateAccess(t *testing.T) {
	resetServerState(t)

	const goroutines = 50

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			c := &Client{name: "user"}
			addClient(c)
			addMessageToHistory("message")
			removeClient(c)
		}()
	}

	wg.Wait()

	if len(clients) != 0 {
		t.Fatalf("expected 0 clients after add+remove, got %d", len(clients))
	}
	if len(messageHistory) != goroutines {
		t.Fatalf("expected %d history entries, got %d", goroutines, len(messageHistory))
	}
}
