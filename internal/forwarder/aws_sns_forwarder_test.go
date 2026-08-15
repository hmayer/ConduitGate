package forwarder

import (
	"conduitgate/internal/config"
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type mockSNSAPI struct {
	SNSAPI
	publishFunc func(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

func (m *mockSNSAPI) Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
	return m.publishFunc(ctx, params, optFns...)
}

func TestSNSForwarder_Forward(t *testing.T) {
	tests := []struct {
		name        string
		destination config.Destination
		payload     []byte
		mockPublish func(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
		wantErr     bool
	}{
		{
			name: "successful publish",
			destination: config.Destination{
				Protocol: "sns",
				Region:   "us-east-1",
				TopicARN: "arn:aws:sns:us-east-1:123456789012:MyTopic",
			},
			payload: []byte("test-message"),
			mockPublish: func(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
				if *params.TopicArn != "arn:aws:sns:us-east-1:123456789012:MyTopic" {
					t.Errorf("Expected TopicArn %s, got %s", "arn:aws:sns:us-east-1:123456789012:MyTopic", *params.TopicArn)
				}
				if *params.Message != "test-message" {
					t.Errorf("Expected Message %s, got %s", "test-message", *params.Message)
				}
				return &sns.PublishOutput{}, nil
			},
			wantErr: false,
		},
		{
			name: "unsupported protocol",
			destination: config.Destination{
				Protocol: "http",
			},
			payload: []byte("{}"),
			wantErr: true,
		},
		{
			name: "sns error",
			destination: config.Destination{
				Protocol: "sns",
				Region:   "us-east-1",
				TopicARN: "arn:aws:sns:us-east-1:123456789012:MyTopic",
			},
			payload: []byte("{}"),
			mockPublish: func(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
				return nil, fmt.Errorf("aws error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewSNSForwarder()

			if tt.mockPublish != nil {
				key := fmt.Sprintf("%s-%s", tt.destination.Region, tt.destination.AccessKey)
				f.clients[key] = &mockSNSAPI{publishFunc: tt.mockPublish}
			}

			err := f.Forward(context.Background(), tt.payload, tt.destination)
			if (err != nil) != tt.wantErr {
				t.Errorf("SNSForwarder.Forward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
