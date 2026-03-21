package listener

import (
	"bytes"
	"chevron-router/internal/config"
	"chevron-router/internal/router"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// MockForwarder is a mock implementation of the Forwarder interface
type MockForwarder struct {
	mu       sync.Mutex
	called   int
	payloads [][]byte
	dest     []config.Destination
}

func (m *MockForwarder) Forward(ctx context.Context, payload []byte, destination config.Destination) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called++
	m.payloads = append(m.payloads, payload)
	m.dest = append(m.dest, destination)
	return nil
}

func TestHandleWebhook(t *testing.T) {
	cfg := &config.Config{
		Port: 8080,
		Rules: []config.Rule{
			{
				SourcePath: "/test",
				Destinations: []config.Destination{
					{Protocol: "http", URL: "http://example.com/1"},
					{Protocol: "http", URL: "http://example.com/2"},
				},
			},
		},
	}
	r := router.NewRouter(cfg)
	mockF := &MockForwarder{}
	l := NewListener(cfg, r, mockF)

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		l.HandleWebhook(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("Route Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/unknown", bytes.NewBufferString("payload"))
		w := httptest.NewRecorder()
		l.HandleWebhook(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("Successful Webhook Acceptance", func(t *testing.T) {
		payload := "hello world"
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(payload))
		w := httptest.NewRecorder()

		// Reset mock
		mockF.mu.Lock()
		mockF.called = 0
		mockF.payloads = nil
		mockF.dest = nil
		mockF.mu.Unlock()

		l.HandleWebhook(w, req)

		if w.Code != http.StatusAccepted {
			t.Errorf("Expected status 202, got %d", w.Code)
		}

		// Since forwarding is async, wait a bit
		deadline := time.Now().Add(100 * time.Millisecond)
		for time.Now().Before(deadline) {
			mockF.mu.Lock()
			count := mockF.called
			mockF.mu.Unlock()
			if count == 2 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		mockF.mu.Lock()
		defer mockF.mu.Unlock()
		if mockF.called != 2 {
			t.Errorf("Expected forwarder to be called twice, got %d", mockF.called)
		}

		for _, p := range mockF.payloads {
			if string(p) != payload {
				t.Errorf("Expected payload %s, got %s", payload, string(p))
			}
		}
	})

	t.Run("Payload Read Failure", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", &errorReader{})
		w := httptest.NewRecorder()
		l.HandleWebhook(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
	})
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge // Some error
}
