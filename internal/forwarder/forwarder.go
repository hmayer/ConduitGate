package forwarder

import (
	"conduitgate/internal/config"
	"context"
)

type Forwarder interface {
	Forward(ctx context.Context, payload []byte, destination config.Destination) error
}
