package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chevron-router/internal/config"
	"chevron-router/internal/forwarder"
	"chevron-router/internal/listener"
	"chevron-router/internal/router"
)

func TestIntegration(t *testing.T) {
	// 1. Destination server (where Chevron will forward the webhook)
	receivedPayload := make(chan string, 1)
	receivedAuth := make(chan string, 1)
	receivedHeaders := make(chan http.Header, 1)
	tsDestination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedPayload <- string(body)
		receivedAuth <- r.Header.Get("Authorization")
		receivedHeaders <- r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer tsDestination.Close()

	// 2. Chevron Setup
	cfg := &config.Config{
		Port: 8080,
		Rules: []config.Rule{
			{
				SourcePath: "/incoming",
				Destinations: []config.Destination{
					{
						Protocol: "http",
						URL:      tsDestination.URL,
						Headers: map[string]string{
							"Content-Type":    "application/json",
							"X-Custom-Header": "foobar",
						},
						Username: "user123",
						Password: "pass123",
					},
				},
			},
		},
	}
	r := router.NewRouter(cfg)
	mf := forwarder.NewMultiForwarder()
	mf.Register("http", forwarder.NewHTTPForwarder())
	l := listener.NewListener(cfg, r, mf)

	// 3. Chevron Listener server
	tsChevron := httptest.NewServer(http.HandlerFunc(l.HandleWebhook))
	defer tsChevron.Close()

	// 4. Send a webhook to Chevron
	payload := `{"event": "test", "data": 123}`
	resp, err := http.Post(tsChevron.URL+"/incoming", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatalf("Failed to send webhook to Chevron: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", resp.StatusCode)
	}

	// 5. Wait for the payload to arrive at the destination
	select {
	case received := <-receivedPayload:
		if received != payload {
			t.Errorf("Expected payload %s, got %s", payload, received)
		}
		auth := <-receivedAuth
		if auth == "" {
			t.Error("Expected Authorization header, got none")
		}
		headers := <-receivedHeaders
		if headers.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", headers.Get("Content-Type"))
		}
		if headers.Get("X-Custom-Header") != "foobar" {
			t.Errorf("Expected X-Custom-Header foobar, got %s", headers.Get("X-Custom-Header"))
		}
	case <-time.After(2 * time.Second):
		t.Error("Timed out waiting for forwarded webhook")
	}
}
