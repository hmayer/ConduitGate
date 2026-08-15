package listener

import (
	"conduitgate/internal/config"
	"conduitgate/internal/forwarder"
	"conduitgate/internal/router"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
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
	mux.HandleFunc("/", l.HandleWebhook)

	log.Printf("Starting listener on port %d...", l.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", l.port), mux)
}

func (l *Listener) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	destinations, err := l.router.Match(path)
	if err != nil {
		log.Printf("No route found for path %s: %v", path, err)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read body: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	log.Printf("Routing webhook from %s to %d destinations", path, len(destinations))

	// Clone the payload as it might be necessary for async processing
	payloadCopy := make([]byte, len(payload))
	copy(payloadCopy, payload)

	for _, dest := range destinations {
		destination := dest // capture loop variable
		go func() {
			// Use Background context for async forwarding to avoid it being canceled when the incoming request finishes
			err := l.forwarder.Forward(context.Background(), payloadCopy, destination)
			if err != nil {
				log.Printf("Failed to forward webhook to %s (%s): %v", destination.URL, destination.Protocol, err)
			} else {
				log.Printf("Successfully forwarded webhook to %s (%s)", destination.URL, destination.Protocol)
			}
		}()
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Webhook received and being processed")
}
