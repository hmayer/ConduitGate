package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := `{
		"port": 9090,
		"rules": [
			{
				"sourcePath": "/test",
				"destinations": [
					{
						"protocol": "http",
						"url": "http://example.com",
						"headers": {
							"Content-Type": "application/json",
							"X-Test": "test-value"
						},
						"username": "admin",
						"password": "secret-password"
					},
					{
						"protocol": "rabbitmq",
						"url": "amqp://localhost",
						"exchange": "events",
						"routingKey": "test.key"
					}
				]
			}
		]
	}`
	tmpFile, err := os.CreateTemp("", "routes.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Port)
	}

	if len(cfg.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(cfg.Rules))
	}

	if cfg.Rules[0].SourcePath != "/test" {
		t.Errorf("Expected sourcePath /test, got %s", cfg.Rules[0].SourcePath)
	}
	if len(cfg.Rules[0].Destinations) != 2 {
		t.Errorf("Expected 2 destinations, got %d", len(cfg.Rules[0].Destinations))
	}

	if cfg.Rules[0].Destinations[0].Username != "admin" {
		t.Errorf("Expected username admin, got %s", cfg.Rules[0].Destinations[0].Username)
	}
	if cfg.Rules[0].Destinations[0].Password != "secret-password" {
		t.Errorf("Expected password secret-password, got %s", cfg.Rules[0].Destinations[0].Password)
	}

	if cfg.Rules[0].Destinations[0].Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", cfg.Rules[0].Destinations[0].Headers["Content-Type"])
	}
	if cfg.Rules[0].Destinations[1].Exchange != "events" {
		t.Errorf("Expected exchange events, got %s", cfg.Rules[0].Destinations[1].Exchange)
	}
	if cfg.Rules[0].Destinations[1].RoutingKey != "test.key" {
		t.Errorf("Expected routingKey test.key, got %s", cfg.Rules[0].Destinations[1].RoutingKey)
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, err := LoadConfig("non_existent_file.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}
