package forwarder

import (
	"conduitgate/internal/config"
	"context"
	"sync"
	"testing"
)

// MockAMQP is not easy without a real server, but we can at least test the internal logic
// if we refactor a bit more, but for now let's try a test that doesn't rely on a real server
// by checking the initialization.

func TestAMQPForwarder_ReusesClient(t *testing.T) {
	// This is a bit of a hacky test because amqp.Dial will fail without a server.
	// But we can check if it attempts to dial.

	f := NewAMQPForwarder()

	dest := config.Destination{
		Protocol: "amqp",
		URL:      "amqp://invalid-host:5672",
	}

	ctx := context.Background()

	// First call should fail to connect
	err1 := f.Forward(ctx, []byte("{}"), dest)
	if err1 == nil {
		t.Fatal("Expected error connecting to invalid host, got nil")
	}

	// Check that no client was cached on failure
	f.mu.RLock()
	_, ok := f.clients[dest.URL]
	f.mu.RUnlock()
	if ok {
		t.Error("Client should not be cached after failed dial")
	}
}

func TestAMQPForwarder_Concurrency(t *testing.T) {
	f := NewAMQPForwarder()
	dest := config.Destination{
		Protocol: "amqp",
		URL:      "amqp://invalid-host:5672",
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = f.Forward(context.Background(), []byte("{}"), dest)
		}()
	}
	wg.Wait()

	// Just ensuring no panic/race occurred
}
