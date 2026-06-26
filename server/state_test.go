package server

import (
	"sync"
	"testing"
)

// TestConcurrentSharedStateAccess hammers the shared state from many
// goroutines at once to prove the mutexes prevent data races. 50 goroutines
// is deliberate pressure (not a real client count) — more contention makes a
// missing lock surface faster under -race.
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
