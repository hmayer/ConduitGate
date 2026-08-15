package listener

import (
	"conduitgate/internal/config"
	"conduitgate/internal/forwarder"
	"conduitgate/internal/router"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Listener struct {
	port      int
	router    *router.Router
	forwarder forwarder.Forwarder
}

func NewListener(cfg *config.Config, r *router.Router, f forwarder.Forwarder) *Listener {
	return &Listener{
		port:      cfg.Port,
		router:    r,
		forwarder: f,
	}
}

func (l *Listener) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", l.HandleHealth)
	mux.HandleFunc("/", l.HandleWebhook)

	slog.Info("Starting listener", "port", l.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", l.port), mux)
}

func (l *Listener) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func (l *Listener) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	destinations, err := l.router.Match(path)
	if err != nil {
		slog.Warn("No route found for webhook", "source_path", path, "status", "failure", "error", err.Error())
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body", "source_path", path, "status", "failure", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	slog.Info("Routing webhook", "source_path", path, "destination_count", len(destinations))

	// Clone the payload as it might be necessary for async processing
	payloadCopy := make([]byte, len(payload))
	copy(payloadCopy, payload)

	for _, dest := range destinations {
		destination := dest // capture loop variable
		go func() {
			// Use Background context for async forwarding to avoid it being canceled when the incoming request finishes
			start := time.Now()
			err := l.forwarder.Forward(context.Background(), payloadCopy, destination)
			duration := time.Since(start).Milliseconds()
			if err != nil {
				slog.Error("Webhook forwarding failed",
					"source_path", path,
					"destination", destinationLabel(destination),
					"protocol", destination.Protocol,
					"status", "failure",
					"duration_ms", duration,
					"error", err.Error(),
				)
			} else {
				slog.Info("Webhook forwarded",
					"source_path", path,
					"destination", destinationLabel(destination),
					"protocol", destination.Protocol,
					"status", "success",
					"duration_ms", duration,
				)
			}
		}()
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Webhook received and being processed")
}

func destinationLabel(d config.Destination) string {
	switch d.Protocol {
	case "amqp", "rabbitmq":
		if d.Exchange != "" {
			return d.Exchange
		}
		return d.URL
	case "sqs":
		return d.QueueURL
	case "sns":
		return d.TopicARN
	default: // http, https
		return d.URL
	}
}
