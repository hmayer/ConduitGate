package forwarder

import (
	"chevron-router/internal/config"
	"context"
	"fmt"
)

type MultiForwarder struct {
	forwarders map[string]Forwarder
}

func NewMultiForwarder() *MultiForwarder {
	return &MultiForwarder{
		forwarders: make(map[string]Forwarder),
	}
}

func (m *MultiForwarder) Register(protocol string, f Forwarder) {
	m.forwarders[protocol] = f
}

func (m *MultiForwarder) Forward(ctx context.Context, payload []byte, destination config.Destination) error {
	f, ok := m.forwarders[destination.Protocol]
	if !ok {
		return fmt.Errorf("no forwarder registered for protocol: %s", destination.Protocol)
	}
	return f.Forward(ctx, payload, destination)
}
