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
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SNSAPI interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

type SNSForwarder struct {
	mu      sync.RWMutex
	clients map[string]SNSAPI
}

func NewSNSForwarder() *SNSForwarder {
	return &SNSForwarder{
		clients: make(map[string]SNSAPI),
	}
}

func (f *SNSForwarder) getClient(ctx context.Context, dest config.Destination) (SNSAPI, error) {
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

	client = sns.NewFromConfig(cfg)
	f.clients[key] = client
	return client, nil
}

func (f *SNSForwarder) Forward(ctx context.Context, payload []byte, destination config.Destination) error {
	if destination.Protocol != "sns" {
		return fmt.Errorf("unsupported protocol: %s", destination.Protocol)
	}

	client, err := f.getClient(ctx, destination)
	if err != nil {
		return err
	}

	_, err = client.Publish(ctx, &sns.PublishInput{
		Message:  aws.String(string(payload)),
		TopicArn: aws.String(destination.TopicARN),
	})

	if err != nil {
		return fmt.Errorf("failed to publish to SNS: %w", err)
	}

	log.Printf("Successfully published to SNS Topic: %s", destination.TopicARN)
	return nil
}
