package forwarder

import (
	"conduitgate/internal/config"
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type mockSQSAPI struct {
	SQSAPI
	sendMessageFunc func(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

func (m *mockSQSAPI) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	return m.sendMessageFunc(ctx, params, optFns...)
}

func TestSQSForwarder_Forward(t *testing.T) {
	tests := []struct {
		name            string
		destination     config.Destination
		payload         []byte
		mockSendMessage func(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
		wantErr         bool
	}{
		{
			name: "successful send",
			destination: config.Destination{
				Protocol: "sqs",
				Region:   "us-east-1",
				QueueURL: "https://sqs.us-east-1.amazonaws.com/123456789012/MyQueue",
			},
			payload: []byte("test-message"),
			mockSendMessage: func(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
				if *params.QueueUrl != "https://sqs.us-east-1.amazonaws.com/123456789012/MyQueue" {
					t.Errorf("Expected QueueUrl %s, got %s", "https://sqs.us-east-1.amazonaws.com/123456789012/MyQueue", *params.QueueUrl)
				}
				if *params.MessageBody != "test-message" {
					t.Errorf("Expected MessageBody %s, got %s", "test-message", *params.MessageBody)
				}
				return &sqs.SendMessageOutput{}, nil
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
			name: "sqs error",
			destination: config.Destination{
				Protocol: "sqs",
				Region:   "us-east-1",
				QueueURL: "https://sqs.us-east-1.amazonaws.com/123456789012/MyQueue",
			},
			payload: []byte("{}"),
			mockSendMessage: func(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
				return nil, fmt.Errorf("aws error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewSQSForwarder()

			if tt.mockSendMessage != nil {
				key := fmt.Sprintf("%s-%s", tt.destination.Region, tt.destination.AccessKey)
				f.clients[key] = &mockSQSAPI{sendMessageFunc: tt.mockSendMessage}
			}

			err := f.Forward(context.Background(), tt.payload, tt.destination)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQSForwarder.Forward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
