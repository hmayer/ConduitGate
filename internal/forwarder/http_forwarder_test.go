package forwarder

import (
	"chevron-router/internal/config"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPForwarder_Forward(t *testing.T) {
	tests := []struct {
		name        string
		destination config.Destination
		payload     []byte
		handler     http.HandlerFunc
		wantErr     bool
	}{
		{
			name: "successful post",
			destination: config.Destination{
				Protocol: "http",
				Headers:  map[string]string{"X-Test": "value"},
			},
			payload: []byte(`{"test":"data"}`),
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST method, got %s", r.Method)
				}
				if r.Header.Get("X-Test") != "value" {
					t.Errorf("Expected header X-Test: value, got %s", r.Header.Get("X-Test"))
				}
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name: "unsupported protocol",
			destination: config.Destination{
				Protocol: "ftp",
			},
			payload: []byte("{}"),
			wantErr: true,
		},
		{
			name: "server error",
			destination: config.Destination{
				Protocol: "http",
			},
			payload: []byte("{}"),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
		{
			name: "basic auth",
			destination: config.Destination{
				Protocol: "http",
				Username: "user",
				Password: "pass",
			},
			payload: []byte("{}"),
			handler: func(w http.ResponseWriter, r *http.Request) {
				username, password, ok := r.BasicAuth()
				if !ok || username != "user" || password != "pass" {
					t.Errorf("Expected basic auth user:pass, got ok:%v, user:%s, pass:%s", ok, username, password)
				}
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewHTTPForwarder()

			if tt.handler != nil {
				server := httptest.NewServer(tt.handler)
				defer server.Close()
				tt.destination.URL = server.URL
			} else if tt.destination.URL == "" && tt.destination.Protocol == "http" {
				tt.destination.URL = "http://localhost:12345" // Invalid address
			}

			err := f.Forward(context.Background(), tt.payload, tt.destination)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPForwarder.Forward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
