package forwarder

import (
	"conduitgate/internal/config"
	"context"
	"errors"
	"testing"
)

type mockForwarder struct {
	payload     []byte
	destination config.Destination
	err         error
	called      bool
}

func (m *mockForwarder) Forward(ctx context.Context, payload []byte, destination config.Destination) error {
	m.called = true
	m.payload = payload
	m.destination = destination
	return m.err
}

func TestMultiForwarder_Forward(t *testing.T) {
	m := NewMultiForwarder()

	httpMock := &mockForwarder{}
	amqpMock := &mockForwarder{}

	m.Register("http", httpMock)
	m.Register("amqp", amqpMock)

	ctx := context.Background()
	payload := []byte("test-payload")

	// Test HTTP routing
	destHTTP := config.Destination{Protocol: "http", URL: "http://example.com"}
	err := m.Forward(ctx, payload, destHTTP)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !httpMock.called {
		t.Error("HTTP forwarder was not called")
	}
	if string(httpMock.payload) != "test-payload" {
		t.Errorf("Expected payload 'test-payload', got %s", httpMock.payload)
	}

	// Test AMQP routing
	destAMQP := config.Destination{Protocol: "amqp", URL: "amqp://localhost"}
	err = m.Forward(ctx, payload, destAMQP)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !amqpMock.called {
		t.Error("AMQP forwarder was not called")
	}

	// Test missing protocol
	destMissing := config.Destination{Protocol: "sns"}
	err = m.Forward(ctx, payload, destMissing)
	if err == nil {
		t.Error("Expected error for missing protocol, got nil")
	}

	// Test error propagation
	errorMock := &mockForwarder{err: errors.New("failed")}
	m.Register("error", errorMock)
	err = m.Forward(ctx, payload, config.Destination{Protocol: "error"})
	if err == nil || err.Error() != "failed" {
		t.Errorf("Expected error 'failed', got %v", err)
	}
}
