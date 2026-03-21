package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Destination struct {
	Protocol   string            `json:"protocol"`
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers,omitempty"`
	Username   string            `json:"username,omitempty"`
	Password   string            `json:"password,omitempty"`
	Exchange   string            `json:"exchange,omitempty"` // Added for AMQP/RabbitMQ
	RoutingKey string            `json:"routingKey,omitempty"`
	Region     string            `json:"region,omitempty"`    // Added for AWS
	AccessKey  string            `json:"accessKey,omitempty"` // Added for AWS
	SecretKey  string            `json:"secretKey,omitempty"` // Added for AWS
	TopicARN   string            `json:"topicARN,omitempty"`  // Added for AWS SNS
	QueueURL   string            `json:"queueURL,omitempty"`  // Added for AWS SQS
}

type Rule struct {
	SourcePath   string        `json:"sourcePath"`
	Destinations []Destination `json:"destinations"` // Changed to support many destinations
}

type Config struct {
	Port  int    `json:"port"`
	Rules []Rule `json:"rules"`
}

func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("could not decode config file: %w", err)
	}

	return &cfg, nil
}
