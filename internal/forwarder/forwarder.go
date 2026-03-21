package forwarder

import (
	"chevron-router/internal/config"
	"context"
)

type Forwarder interface {
	Forward(ctx context.Context, payload []byte, destination config.Destination) error
}
