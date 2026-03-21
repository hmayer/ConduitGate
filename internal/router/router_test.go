package router

import (
	"chevron-router/internal/config"
	"testing"
)

func TestRouter_Match(t *testing.T) {
	cfg := &config.Config{
		Rules: []config.Rule{
			{
				SourcePath: "/webhook",
				Destinations: []config.Destination{
					{Protocol: "http", URL: "http://example.com/dest1"},
					{Protocol: "http", URL: "http://example.com/dest2"},
				},
			},
			{
				SourcePath: "/other",
				Destinations: []config.Destination{
					{Protocol: "amqp", URL: "amqp://rabbitmq:5672"},
				},
			},
		},
	}

	r := NewRouter(cfg)

	tests := []struct {
		name          string
		path          string
		expectedLen   int
		expectedError bool
	}{
		{
			name:          "match-webhook",
			path:          "/webhook",
			expectedLen:   2,
			expectedError: false,
		},
		{
			name:          "match-other",
			path:          "/other",
			expectedLen:   1,
			expectedError: false,
		},
		{
			name:          "no-match",
			path:          "/notfound",
			expectedLen:   0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destinations, err := r.Match(tt.path)
			if (err != nil) != tt.expectedError {
				t.Errorf("Match() error = %v, expectedError %v", err, tt.expectedError)
				return
			}
			if len(destinations) != tt.expectedLen {
				t.Errorf("Match() got %d destinations, expected %d", len(destinations), tt.expectedLen)
			}
		})
	}
}
