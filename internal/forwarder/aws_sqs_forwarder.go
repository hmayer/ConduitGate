package forwarder

import (
	"conduitgate/internal/config"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSAPI interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

type SQSForwarder struct {
	mu      sync.RWMutex
	clients map[string]SQSAPI
}

func NewSQSForwarder() *SQSForwarder {
	return &SQSForwarder{
		clients: make(map[string]SQSAPI),
	}
}

func (f *SQSForwarder) getClient(ctx context.Context, dest config.Destination) (SQSAPI, error) {
	key := fmt.Sprintf("%s-%s", dest.Region, dest.AccessKey)

	f.mu.RLock()
	client, ok := f.clients[key]
	f.mu.RUnlock()

	if ok {
		return client, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Check again
	if client, ok := f.clients[key]; ok {
		return client, nil
	}

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(dest.Region),
	}

	if dest.AccessKey != "" && dest.SecretKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(dest.AccessKey, dest.SecretKey, ""),
		))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	client = sqs.NewFromConfig(cfg)
	f.clients[key] = client
	return client, nil
}

func (f *SQSForwarder) Forward(ctx context.Context, payload []byte, destination config.Destination) error {
	if destination.Protocol != "sqs" {
		return fmt.Errorf("unsupported protocol: %s", destination.Protocol)
	}

	client, err := f.getClient(ctx, destination)
	if err != nil {
		return err
	}

	_, err = client.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(string(payload)),
		QueueUrl:    aws.String(destination.QueueURL),
	})

	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %w", err)
	}

	log.Printf("Successfully sent message to SQS Queue: %s", destination.QueueURL)
	return nil
}
