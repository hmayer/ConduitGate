package forwarder

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"chevron-router/internal/config"
)

type HTTPForwarder struct {
	client *http.Client
}

func NewHTTPForwarder() *HTTPForwarder {
	return &HTTPForwarder{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (f *HTTPForwarder) Forward(ctx context.Context, payload []byte, destination config.Destination) error {
	if destination.Protocol != "http" && destination.Protocol != "https" {
		return fmt.Errorf("unsupported protocol: %s", destination.Protocol)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, destination.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range destination.Headers {
		req.Header.Set(k, v)
	}

	if destination.Username != "" || destination.Password != "" {
		req.SetBasicAuth(destination.Username, destination.Password)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
